package commands

import (
	"fmt"
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
		return nil, err
	}

	h.DB.Preload("User").Preload("Tier").Where("id = ?", partner.ID).First(&partner)
	return &partner, nil
}
