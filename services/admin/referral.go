package admin

import (
	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/services"
	"xmeta-partner/structs"
)

type ReferralService struct {
	services.BaseService
}

// List returns paginated referrals for admin views.
//
// One endpoint serves both surfaces:
//   - global /partner/referrals page → no PartnerID filter
//   - per-partner drawer tab → PartnerID set
//
// Default behavior hides historical rows (ended_at IS NOT NULL). Pass
// IncludeHistorical=true to also surface past relationships — useful when
// auditing why a user shows up under a different partner now.
func (s *ReferralService) List(params structs.AdminReferralListParams) (structs.PaginationResponse, error) {
	pInput := services.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := s.DB.Model(&database.Referral{}).
		Preload("ReferredUser").
		Preload("ReferralLink").
		Preload("Partner.User").
		Preload("Partner.Tier")

	orm = common.Equal(orm, "partner_id", params.PartnerID)
	orm = common.Equal(orm, "status", params.Status)

	if !params.IncludeHistorical {
		orm = orm.Where("ended_at IS NULL")
	}

	if params.Query != "" {
		q := "%" + params.Query + "%"
		orm = orm.
			Joins("LEFT JOIN users ON users.id = referrals.referred_user_id").
			Where("users.email ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ?", q, q, q)
	}

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var referrals []database.Referral
	if err := orm.
		Order("referrals.created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&referrals).Error; err != nil {
		return structs.PaginationResponse{}, err
	}

	return structs.PaginationResponse{Total: total, Items: referrals}, nil
}

// Detail returns a single referral with the related user + partner info,
// plus the user's commission history under this partner so the admin
// drawer can show "what trades earned this partner money from this user".
func (s *ReferralService) Detail(id string) (map[string]interface{}, error) {
	var referral database.Referral
	if err := s.DB.
		Preload("ReferredUser").
		Preload("ReferralLink").
		Preload("Partner.User").
		Preload("Partner.Tier").
		Where("id = ?", id).
		First(&referral).Error; err != nil {
		return nil, err
	}

	var commissions []database.Commission
	s.DB.
		Where("partner_id = ? AND referred_user_id = ?", referral.PartnerID, referral.ReferredUserID).
		Order("trade_date desc").
		Limit(50).
		Find(&commissions)

	var totalEarned float64
	s.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND referred_user_id = ?", referral.PartnerID, referral.ReferredUserID).
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&totalEarned)

	var totalVolume float64
	s.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND referred_user_id = ?", referral.PartnerID, referral.ReferredUserID).
		Select("COALESCE(SUM(trade_amount), 0)").
		Scan(&totalVolume)

	return map[string]interface{}{
		"referral":    referral,
		"commissions": commissions,
		"totalEarned": totalEarned,
		"totalVolume": totalVolume,
	}, nil
}
