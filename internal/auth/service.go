package auth

import (
	"encoding/json"
	"fmt"

	"xmeta-partner/database"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{DB: db}
}

type AuthStatusResponse struct {
	User        *database.User               `json:"user"`
	Partner     *database.Partner            `json:"partner"`
	Application *database.PartnerApplication `json:"application"`
}

func (s *Service) GetAuthStatus(user *database.User) (AuthStatusResponse, error) {
	resp := AuthStatusResponse{User: user}

	var partner database.Partner
	if err := s.DB.Preload("Tier").Where("user_id = ?", user.ID).First(&partner).Error; err == nil {
		partner.User = user
		resp.Partner = &partner
	}

	var app database.PartnerApplication
	if err := s.DB.Where("user_id = ?", user.ID).
		Order("created_at DESC").First(&app).Error; err == nil {
		resp.Application = &app
	}

	return resp, nil
}

func (s *Service) Signup(params structs.PartnerSignupParams) (database.PartnerApplication, error) {
	if params.UserID == "" {
		return database.PartnerApplication{}, fmt.Errorf("userId is required")
	}

	var user database.User
	if err := s.DB.Where("id = ?", params.UserID).First(&user).Error; err != nil {
		return database.PartnerApplication{}, fmt.Errorf("user not found")
	}
	if user.KycLevel < 1 {
		return database.PartnerApplication{}, fmt.Errorf("KYC verification required before applying as a partner")
	}

	var existingCount int64
	s.DB.Model(&database.PartnerApplication{}).
		Where("user_id = ? AND status IN ?", params.UserID, []database.ApplicationStatus{
			database.ApplicationStatusPending,
			database.ApplicationStatusApproved,
		}).
		Count(&existingCount)
	if existingCount > 0 {
		return database.PartnerApplication{}, fmt.Errorf("an application is already pending or approved for this user")
	}

	var partnerCount int64
	s.DB.Model(&database.Partner{}).Where("user_id = ?", params.UserID).Count(&partnerCount)
	if partnerCount > 0 {
		return database.PartnerApplication{}, fmt.Errorf("a partner account already exists for this user")
	}

	app := database.PartnerApplication{
		UserID:        params.UserID,
		CompanyName:   params.CompanyName,
		Website:       params.Website,
		SocialMedia:   params.SocialMedia,
		AudienceSize:  params.AudienceSize,
		PromotionPlan: params.PromotionPlan,
		Status:        database.ApplicationStatusPending,
	}

	if err := s.DB.Create(&app).Error; err != nil {
		return database.PartnerApplication{}, err
	}

	return app, nil
}

func (s *Service) GetInfo(partnerID string) (database.Partner, error) {
	var partner database.Partner
	if err := s.DB.Preload("User").Preload("Tier").Where("id = ?", partnerID).First(&partner).Error; err != nil {
		return database.Partner{}, err
	}
	return partner, nil
}

func (s *Service) UpdateProfile(partnerID string, params structs.PartnerProfileUpdateParams) (database.Partner, error) {
	var partner database.Partner
	if err := s.DB.Where("id = ?", partnerID).First(&partner).Error; err != nil {
		return database.Partner{}, err
	}

	updates := map[string]interface{}{}
	if params.CompanyName != nil {
		updates["company_name"] = *params.CompanyName
	}
	if params.Website != nil {
		updates["website"] = *params.Website
	}
	if params.SocialMedia != nil {
		bytes, err := json.Marshal(params.SocialMedia)
		if err != nil {
			return database.Partner{}, fmt.Errorf("marshal socialMedia: %w", err)
		}
		updates["social_media"] = string(bytes)
	}

	if len(updates) > 0 {
		if err := s.DB.Model(&partner).Updates(updates).Error; err != nil {
			return database.Partner{}, err
		}
	}

	return s.GetInfo(partnerID)
}

func (s *Service) GetTierDetails(partnerID string) (map[string]interface{}, error) {
	var partner database.Partner
	if err := s.DB.Preload("Tier").Where("id = ?", partnerID).First(&partner).Error; err != nil {
		return nil, err
	}

	var tiers []database.PartnerTier
	if err := s.DB.Order("level asc").Find(&tiers).Error; err != nil {
		return nil, err
	}

	var totalVolume float64
	s.DB.Model(&database.Commission{}).
		Where("partner_id = ?", partnerID).
		Select("COALESCE(SUM(trade_amount), 0)").
		Scan(&totalVolume)

	var activeClients int64
	s.DB.Model(&database.Referral{}).
		Where("partner_id = ? AND status = ? AND ended_at IS NULL", partnerID, database.ReferralStatusActive).
		Count(&activeClients)

	result := map[string]interface{}{
		"currentTier":   partner.Tier,
		"activeClients": activeClients,
		"totalVolume":   totalVolume,
	}

	if partner.Tier != nil {
		for _, tier := range tiers {
			if tier.Level > partner.Tier.Level {
				activeClientsProgress := float64(0)
				if tier.MinActiveClients > 0 {
					activeClientsProgress = float64(activeClients) / float64(tier.MinActiveClients)
					if activeClientsProgress > 1 {
						activeClientsProgress = 1
					}
				}
				volumeProgress := float64(0)
				if tier.MinVolume > 0 {
					volumeProgress = totalVolume / tier.MinVolume
					if volumeProgress > 1 {
						volumeProgress = 1
					}
				}

				result["nextTier"] = tier
				result["activeClientsProgress"] = activeClientsProgress
				result["volumeProgress"] = volumeProgress
				break
			}
		}
	}

	return result, nil
}

func (s *Service) TrackClick(code string, ipAddress string, userAgent string) (map[string]interface{}, error) {
	var link database.ReferralLink
	if err := s.DB.Where("code = ? AND is_active = ?", code, true).First(&link).Error; err != nil {
		return nil, fmt.Errorf("referral link not found")
	}

	s.DB.Model(&link).UpdateColumn("clicks", s.DB.Raw("clicks + 1"))

	return map[string]interface{}{
		"url":  link.URL,
		"code": link.Code,
	}, nil
}
