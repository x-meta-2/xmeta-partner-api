package commands

import (
	"fmt"
	"log"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/partner/domain"
	"xmeta-partner/internal/partner/port"
	"xmeta-partner/utils"

	"gorm.io/gorm"
)

type ApproveApplicationHandler struct {
	DB       *gorm.DB
	Apps     port.ApplicationRepo
	Tiers    port.TierRepo
	Partners port.PartnerRepo
}

func (h *ApproveApplicationHandler) Handle(applicationID string, adminID string) (*database.Partner, error) {
	var partner database.Partner
	log.Printf("[Email] approve application flow started applicationId=%s adminId=%s", applicationID, adminID)

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var app database.PartnerApplication
		if err := tx.Where("id = ? AND status = ?", applicationID, domain.ApplicationPending).First(&app).Error; err != nil {
			return domain.ErrApplicationNotFound
		}

		now := time.Now()
		app.Status = database.ApplicationStatusApproved
		app.ReviewedBy = &adminID
		app.ReviewedAt = &now
		if err := tx.Save(&app).Error; err != nil {
			return err
		}

		var defaultTier database.PartnerTier
		if err := tx.Where("is_default = ?", true).First(&defaultTier).Error; err != nil {
			return domain.ErrDefaultTierMissing
		}

		var referralCode string
		for attempt := 0; attempt < 5; attempt++ {
			candidate, err := utils.GenerateReferralCode()
			if err != nil {
				return err
			}
			var linkCount, partnerCount int64
			tx.Model(&database.ReferralLink{}).Where("code = ?", candidate).Count(&linkCount)
			tx.Model(&database.Partner{}).Where("referral_code = ?", candidate).Count(&partnerCount)
			if linkCount == 0 && partnerCount == 0 {
				referralCode = candidate
				break
			}
		}
		if referralCode == "" {
			return domain.ErrReferralCodeCollision
		}

		partner = database.Partner{
			UserID:       app.UserID,
			CompanyName:  app.CompanyName,
			Website:      app.Website,
			SocialMedia:  app.SocialMedia,
			TierID:       defaultTier.ID,
			Status:       database.PartnerStatusActive,
			ReferralCode: referralCode,
		}
		if err := tx.Create(&partner).Error; err != nil {
			return err
		}

		defaultLink := database.ReferralLink{
			PartnerID: partner.ID,
			Code:      referralCode,
			URL:       fmt.Sprintf("https://x-meta.com/?ref=%s", referralCode),
			IsActive:  true,
		}
		return tx.Create(&defaultLink).Error
	})

	if err != nil {
		log.Printf("[Email] approve application flow stopped applicationId=%s error=%v", applicationID, err)
		return nil, err
	}

	if err := h.DB.Preload("User").Preload("Tier").Where("id = ?", partner.ID).First(&partner).Error; err != nil {
		log.Printf("[Email] approved partner reload failed partnerId=%s userId=%s error=%v", partner.ID, partner.UserID, err)
		return nil, err
	}
	log.Printf("[Email] approved partner reloaded partnerId=%s userId=%s userLoaded=%t", partner.ID, partner.UserID, partner.User != nil)

	go func() {
		if es := utils.GetEmailService(); es != nil && partner.User != nil {
			var app database.PartnerApplication
			if err := h.DB.Where("user_id = ? AND status = ?", partner.UserID, database.ApplicationStatusApproved).
				Order("updated_at DESC").First(&app).Error; err != nil {
				log.Printf("ERROR: could not load application for approved email: %v", err)
				return
			}
			locale := app.Locale
			if locale == "" {
				locale = "mn"
			}
			log.Printf("[Email] partner approved email trigger partnerId=%s userId=%s locale=%s", partner.ID, partner.UserID, locale)
			es.SendPartnerApprovedEmail(partner.User.Email, locale)
		} else {
			log.Printf("[Email] partner approved email skipped partnerId=%s userId=%s reason=email_service_or_user_missing", partner.ID, partner.UserID)
		}
	}()

	return &partner, nil
}
