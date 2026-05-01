package admin

import (
	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/services"
	"xmeta-partner/structs"
)

type PartnerService struct {
	services.BaseService
}

// List returns paginated partners with tier/status/email filter, preload Tier
func (s *PartnerService) List(params structs.PartnerListParams) (structs.PaginationResponse, error) {
	pInput := services.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := s.DB.Model(&database.Partner{}).Preload("User").Preload("Tier")
	orm = common.Equal(orm, "status", params.Status)
	orm = common.Equal(orm, "tier_id", params.TierID)

	if params.Query != "" {
		q := "%" + params.Query + "%"
		orm = orm.
			Joins("LEFT JOIN users ON users.id = partners.user_id").
			Where(
				"users.email ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ? OR partners.company_name ILIKE ? OR partners.referral_code ILIKE ?",
				q, q, q, q, q,
			)
	}

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var partners []database.Partner
	if err := orm.
		Order("created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&partners).Error; err != nil {
		return structs.PaginationResponse{}, err
	}

	return structs.PaginationResponse{Total: total, Items: partners}, nil
}

// Detail returns a partner with tier, recent referrals, and recent commissions
func (s *PartnerService) Detail(id string) (map[string]interface{}, error) {
	var partner database.Partner
	if err := s.DB.Preload("User").Preload("Tier").Where("id = ?", id).First(&partner).Error; err != nil {
		return nil, err
	}

	var recentReferrals []database.Referral
	s.DB.Where("partner_id = ?", id).
		Order("created_at desc").
		Limit(10).
		Find(&recentReferrals)

	var recentCommissions []database.Commission
	s.DB.Where("partner_id = ?", id).
		Order("created_at desc").
		Limit(10).
		Find(&recentCommissions)

	// All referral links the partner has created (primary code + custom
	// codes from the partner-portal). The list is small (max 5) so we
	// return all rows.
	var referralLinks []database.ReferralLink
	s.DB.Where("partner_id = ?", id).
		Order("created_at asc").
		Find(&referralLinks)

	var totalVolume float64
	s.DB.Model(&database.Commission{}).
		Where("partner_id = ?", id).
		Select("COALESCE(SUM(trade_amount), 0)").
		Scan(&totalVolume)

	var pendingCommissions float64
	s.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND status = ?", id, "pending").
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&pendingCommissions)

	return map[string]interface{}{
		"partner":            partner,
		"recentReferrals":    recentReferrals,
		"recentCommissions":  recentCommissions,
		"referralLinks":      referralLinks,
		"totalVolume":        totalVolume,
		"pendingCommissions": pendingCommissions,
	}, nil
}

func (s *PartnerService) UpdateTier(id string, tierID string) (database.Partner, error) {
	var tier database.PartnerTier
	if err := s.DB.Where("id = ?", tierID).First(&tier).Error; err != nil {
		return database.Partner{}, err
	}

	if err := s.DB.Model(&database.Partner{}).Where("id = ?", id).Update("tier_id", tierID).Error; err != nil {
		return database.Partner{}, err
	}

	var partner database.Partner
	s.DB.Preload("User").Preload("Tier").Where("id = ?", id).First(&partner)
	return partner, nil
}

func (s *PartnerService) UpdateStatus(id string, status string) (database.Partner, error) {
	if err := s.DB.Model(&database.Partner{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return database.Partner{}, err
	}

	var partner database.Partner
	s.DB.Preload("User").Preload("Tier").Where("id = ?", id).First(&partner)
	return partner, nil
}
