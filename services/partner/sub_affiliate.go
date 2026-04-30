package partner

import (
	"fmt"
	"time"

	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/services"
	"xmeta-partner/structs"
	"xmeta-partner/utils"
)

type SubAffiliateService struct {
	services.BaseService
}

// List returns sub-affiliates (partners where parent_id = partnerID) with their referral/earning stats
func (s *SubAffiliateService) List(partnerID string, params structs.SubAffiliateListParams) (structs.PaginationResponse, error) {
	pInput := services.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := s.DB.Model(&database.Partner{}).Where("parent_id = ?", partnerID)
	orm = common.Equal(orm, "status", params.Status)

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var subPartners []database.Partner
	if err := orm.
		Preload("Tier").
		Order("created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&subPartners).Error; err != nil {
		return structs.PaginationResponse{}, err
	}

	type SubAffiliateItem struct {
		database.Partner
		OverrideEarnings float64 `json:"overrideEarnings"`
	}

	var items []SubAffiliateItem
	for _, sp := range subPartners {
		var overrideEarnings float64
		s.DB.Model(&database.Commission{}).
			Where("partner_id = ? AND is_override = ? AND override_partner_id = ?", partnerID, true, sp.ID).
			Select("COALESCE(SUM(commission_amount), 0)").
			Scan(&overrideEarnings)

		items = append(items, SubAffiliateItem{
			Partner:          sp,
			OverrideEarnings: overrideEarnings,
		})
	}

	return structs.PaginationResponse{Total: total, Items: items}, nil
}

// Invite creates a sub-affiliate invitation with generated code and 7-day expiry
func (s *SubAffiliateService) Invite(partnerID string, params structs.SubAffiliateInviteParams) (database.SubAffiliateInvite, error) {
	var existingCount int64
	s.DB.Model(&database.SubAffiliateInvite{}).
		Where("partner_id = ? AND email = ? AND status = ?", partnerID, params.Email, "pending").
		Count(&existingCount)
	if existingCount > 0 {
		return database.SubAffiliateInvite{}, fmt.Errorf("an invitation to this email is already pending")
	}

	code, err := utils.GenerateReferralCode()
	if err != nil {
		return database.SubAffiliateInvite{}, err
	}

	overrideRate := 0.10
	if params.OverrideRate != nil {
		overrideRate = *params.OverrideRate
	}

	expiresAt := time.Now().AddDate(0, 0, 7)
	invite := database.SubAffiliateInvite{
		PartnerID:    partnerID,
		Email:        params.Email,
		InviteCode:   code,
		Status:       "pending",
		OverrideRate: overrideRate,
		ExpiresAt:    &expiresAt,
	}

	if err := s.DB.Create(&invite).Error; err != nil {
		return database.SubAffiliateInvite{}, err
	}

	return invite, nil
}

func (s *SubAffiliateService) InviteDetail(partnerID string, code string) (database.SubAffiliateInvite, error) {
	var invite database.SubAffiliateInvite
	if err := s.DB.Preload("Partner").
		Where("partner_id = ? AND invite_code = ?", partnerID, code).
		First(&invite).Error; err != nil {
		return database.SubAffiliateInvite{}, fmt.Errorf("invite not found")
	}
	return invite, nil
}

// Stats returns sub-affiliate statistics: total sub-affiliates, active count, total override earned
func (s *SubAffiliateService) Stats(partnerID string, params structs.SubAffiliateStatsParams) (map[string]interface{}, error) {
	var totalInvites int64
	s.DB.Model(&database.SubAffiliateInvite{}).Where("partner_id = ?", partnerID).Count(&totalInvites)

	var acceptedInvites int64
	s.DB.Model(&database.SubAffiliateInvite{}).Where("partner_id = ? AND status = ?", partnerID, "accepted").Count(&acceptedInvites)

	var pendingInvites int64
	s.DB.Model(&database.SubAffiliateInvite{}).Where("partner_id = ? AND status = ?", partnerID, "pending").Count(&pendingInvites)

	var totalSubAffiliates int64
	s.DB.Model(&database.Partner{}).Where("parent_id = ?", partnerID).Count(&totalSubAffiliates)

	var activeSubAffiliates int64
	s.DB.Model(&database.Partner{}).Where("parent_id = ? AND status = ?", partnerID, "active").Count(&activeSubAffiliates)

	var overrideEarnings float64
	s.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND is_override = ?", partnerID, true).
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&overrideEarnings)

	return map[string]interface{}{
		"totalInvites":        totalInvites,
		"acceptedInvites":     acceptedInvites,
		"pendingInvites":      pendingInvites,
		"totalSubAffiliates":  totalSubAffiliates,
		"activeSubAffiliates": activeSubAffiliates,
		"overrideEarnings":    overrideEarnings,
	}, nil
}
