package admin

import (
	"fmt"
	"time"

	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/services"
	"xmeta-partner/structs"
	"xmeta-partner/utils"

	"gorm.io/gorm"
)

type ApplicationService struct {
	services.BaseService
}

// List returns paginated partner applications with status/email filter
func (s *ApplicationService) List(params structs.ApplicationListParams) (structs.PaginationResponse, error) {
	pInput := services.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := s.DB.Model(&database.PartnerApplication{}).Preload("User")
	orm = common.Equal(orm, "status", params.Status)

	if params.Query != "" {
		q := "%" + params.Query + "%"
		orm = orm.
			Joins("LEFT JOIN users ON users.id = partner_applications.user_id").
			Where(
				"users.email ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ? OR partner_applications.company_name ILIKE ?",
				q, q, q, q,
			)
	}

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var applications []database.PartnerApplication
	if err := orm.
		Order("created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&applications).Error; err != nil {
		return structs.PaginationResponse{}, err
	}

	return structs.PaginationResponse{Total: total, Items: applications}, nil
}

func (s *ApplicationService) Detail(id string) (database.PartnerApplication, error) {
	var app database.PartnerApplication
	if err := s.DB.Where("id = ?", id).First(&app).Error; err != nil {
		return database.PartnerApplication{}, err
	}
	return app, nil
}

// Approve approves a partner application, creates Partner record with referral_code and default tier.
// Uses a GORM transaction to ensure atomicity.
func (s *ApplicationService) Approve(id string, adminID string) (database.Partner, error) {
	var partner database.Partner

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var app database.PartnerApplication
		if err := tx.Where("id = ? AND status = ?", id, database.ApplicationStatusPending).First(&app).Error; err != nil {
			return fmt.Errorf("application not found or already reviewed")
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
			return fmt.Errorf("no default partner tier configured — set one in Partner Config first")
		}

		// Generate a unique 7-char referral code. Retry up to 5 times
		// against the live unique indexes (`partners.referral_code` and
		// `referral_links.code` share the same namespace) so a one-in-78B
		// collision never blocks an approve.
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
			return fmt.Errorf("could not generate a unique referral code; retry approval")
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

		// Every newly approved partner starts with one default referral
		// link. They can add up to two more (max 3 total) from Settings.
		defaultLink := database.ReferralLink{
			PartnerID: partner.ID,
			Code:      referralCode,
			URL:       fmt.Sprintf("https://x-meta.com/?ref=%s", referralCode),
			IsActive:  true,
		}
		if err := tx.Create(&defaultLink).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return database.Partner{}, err
	}

	s.DB.Preload("User").Preload("Tier").Where("id = ?", partner.ID).First(&partner)
	return partner, nil
}

func (s *ApplicationService) Reject(id string, adminID string, params structs.ApplicationReviewParams) (database.PartnerApplication, error) {
	var app database.PartnerApplication
	if err := s.DB.Where("id = ? AND status = ?", id, database.ApplicationStatusPending).First(&app).Error; err != nil {
		return database.PartnerApplication{}, fmt.Errorf("application not found or already reviewed")
	}

	now := time.Now()
	app.Status = database.ApplicationStatusRejected
	app.ReviewedBy = &adminID
	app.ReviewedAt = &now
	app.RejectionReason = params.RejectionReason

	if err := s.DB.Save(&app).Error; err != nil {
		return database.PartnerApplication{}, err
	}

	return app, nil
}
