package adapters

import (
	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/internal/partner/app/dto"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type GormPartnerRepo struct {
	DB *gorm.DB
}

func (r *GormPartnerRepo) FindByID(id string) (*database.Partner, error) {
	var partner database.Partner
	if err := r.DB.Preload("User").Preload("Tier").Where("id = ?", id).First(&partner).Error; err != nil {
		return nil, err
	}
	return &partner, nil
}

func (r *GormPartnerRepo) List(params structs.PartnerListParams) ([]database.Partner, int, error) {
	pInput := common.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := r.DB.Model(&database.Partner{}).Preload("User").Preload("Tier")
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
		return nil, 0, err
	}

	return partners, total, nil
}

func (r *GormPartnerRepo) UpdateTier(id string, tierID string) error {
	return r.DB.Model(&database.Partner{}).Where("id = ?", id).Update("tier_id", tierID).Error
}

func (r *GormPartnerRepo) UpdateStatus(id string, status string) error {
	return r.DB.Model(&database.Partner{}).Where("id = ?", id).Update("status", status).Error
}

func (r *GormPartnerRepo) GetDetail(id string) (*dto.PartnerDetail, error) {
	var partner database.Partner
	if err := r.DB.Preload("User").Preload("Tier").Where("id = ?", id).First(&partner).Error; err != nil {
		return nil, err
	}

	var recentReferrals []database.Referral
	r.DB.Where("partner_id = ?", id).Order("created_at desc").Limit(10).Find(&recentReferrals)

	var recentCommissions []database.Commission
	r.DB.Where("partner_id = ?", id).Order("created_at desc").Limit(10).Find(&recentCommissions)

	var referralLinks []database.ReferralLink
	r.DB.Where("partner_id = ?", id).Order("created_at asc").Find(&referralLinks)
	for i := range referralLinks {
		var live int64
		r.DB.Model(&database.Referral{}).
			Where("referral_link_id = ? AND ended_at IS NULL", referralLinks[i].ID).
			Count(&live)
		referralLinks[i].Registrations = int(live)
	}

	var totalVolume float64
	r.DB.Model(&database.Commission{}).
		Where("partner_id = ?", id).
		Select("COALESCE(SUM(trade_amount), 0)").
		Scan(&totalVolume)

	var pendingCommissions float64
	r.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND status = ?", id, "pending").
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&pendingCommissions)

	return &dto.PartnerDetail{
		Partner:            partner,
		RecentReferrals:    recentReferrals,
		RecentCommissions:  recentCommissions,
		ReferralLinks:      referralLinks,
		TotalVolume:        totalVolume,
		PendingCommissions: pendingCommissions,
	}, nil
}
