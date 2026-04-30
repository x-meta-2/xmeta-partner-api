package partner

import (
	"encoding/json"
	"fmt"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/services"
	"xmeta-partner/structs"
)

type AuthService struct {
	services.BaseService
}

// AuthStatusResponse — /auth/status-аас буцаагдах snapshot.
// Frontend энэ response-с хамаарч onboarding эсвэл dashboard руу router хийнэ.
type AuthStatusResponse struct {
	User        *database.User               `json:"user"`
	Partner     *database.Partner            `json:"partner"`
	Application *database.PartnerApplication `json:"application"`
}

// GetAuthStatus — UserAuth middleware-ээр дамжсан user-ийн partner/application
// төлвийг буцаана. Зөвхөн нэг хүн нэг дор нэг application хадгалж болно гэсэн
// таамаг — хэрэв reapply logic нэмбэл "latest application" буцаана.
func (s *AuthService) GetAuthStatus(user *database.User) (AuthStatusResponse, error) {
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

// Signup creates a new partner application.
// The applying user is identified by Cognito sub (UserID); email/name come
// from the shared users table, not this request payload.
func (s *AuthService) Signup(params structs.PartnerSignupParams) (database.PartnerApplication, error) {
	if params.UserID == "" {
		return database.PartnerApplication{}, fmt.Errorf("userId is required")
	}

	// KYC gate — only verified users (kyc_level >= 1) can apply. Enforced on
	// the frontend as well (PartnerGate + onboarding), but server is the
	// ultimate source of truth.
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

func (s *AuthService) GetInfo(partnerID string) (database.Partner, error) {
	var partner database.Partner
	if err := s.DB.Preload("User").Preload("Tier").Where("id = ?", partnerID).First(&partner).Error; err != nil {
		return database.Partner{}, err
	}
	return partner, nil
}

// UpdateProfile updates the partner-owned profile (company name, website,
// social URLs). Identity fields (first/last name, email) live on User
// and are owned by xmeta-monorepo — never edited from here. The original
// PartnerApplication is kept untouched as the review-time snapshot.
func (s *AuthService) UpdateProfile(partnerID string, params structs.PartnerProfileUpdateParams) (database.Partner, error) {
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
		// Updates(map) bypasses the GORM serializer, so the raw map can't
		// be encoded to TEXT by the driver. Marshal here instead so the
		// column gets a valid JSON string.
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

// GetTierDetails returns current tier details with progress to next tier
func (s *AuthService) GetTierDetails(partnerID string) (map[string]interface{}, error) {
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

	result := map[string]interface{}{
		"currentTier":    partner.Tier,
		"totalReferrals": partner.TotalReferrals,
		"totalEarnings":  partner.TotalEarnings,
		"totalVolume":    totalVolume,
	}

	if partner.Tier != nil {
		for _, tier := range tiers {
			if tier.Level > partner.Tier.Level {
				clientsNeeded := tier.MinActiveClients - partner.TotalReferrals
				if clientsNeeded < 0 {
					clientsNeeded = 0
				}
				volumeNeeded := tier.MinVolume - totalVolume
				if volumeNeeded < 0 {
					volumeNeeded = 0
				}

				clientProgress := float64(0)
				if tier.MinActiveClients > 0 {
					clientProgress = float64(partner.TotalReferrals) / float64(tier.MinActiveClients) * 100
					if clientProgress > 100 {
						clientProgress = 100
					}
				}
				volumeProgress := float64(0)
				if tier.MinVolume > 0 {
					volumeProgress = totalVolume / tier.MinVolume * 100
					if volumeProgress > 100 {
						volumeProgress = 100
					}
				}

				overallProgress := (clientProgress + volumeProgress) / 2

				result["nextTier"] = tier
				result["clientsNeeded"] = clientsNeeded
				result["volumeNeeded"] = volumeNeeded
				result["clientProgress"] = clientProgress
				result["volumeProgress"] = volumeProgress
				result["overallProgress"] = overallProgress
				break
			}
		}
	}

	return result, nil
}

func (s *AuthService) TrackClick(code string, ipAddress string, userAgent string) (map[string]interface{}, error) {
	var link database.ReferralLink
	if err := s.DB.Where("code = ? AND is_active = ?", code, true).First(&link).Error; err != nil {
		return nil, fmt.Errorf("referral link not found")
	}

	s.DB.Model(&link).UpdateColumn("clicks", s.DB.Raw("clicks + 1"))

	now := time.Now()
	today := now.Format("2006-01-02")
	// Raw INSERT bypasses GORM auto timestamps; populate created_at /
	// updated_at explicitly so the NOT NULL constraints stay happy.
	s.DB.Exec(
		`INSERT INTO partner_daily_stats (id, created_at, updated_at, partner_id, date, clicks)
		 VALUES (gen_random_uuid(), ?, ?, ?, ?, 1)
		 ON CONFLICT (partner_id, date)
		 DO UPDATE SET clicks = partner_daily_stats.clicks + 1,
		               updated_at = EXCLUDED.updated_at`,
		now, now, link.PartnerID, today,
	)

	return map[string]interface{}{
		"url":  link.URL,
		"code": link.Code,
	}, nil
}
