package adapters

import (
	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/internal/commission/app/dto"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type GormCommissionRepo struct {
	DB *gorm.DB
}

func (r *GormCommissionRepo) List(partnerID string, params structs.CommissionListParams) ([]database.Commission, int, error) {
	pInput := common.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := r.DB.Model(&database.Commission{}).Where("partner_id = ?", partnerID)
	orm = common.Equal(orm, "status", params.Status)
	orm = common.Equal(orm, "referred_user_id", params.ReferredUserID)

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var commissions []database.Commission
	if err := orm.
		Order("created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&commissions).Error; err != nil {
		return nil, 0, err
	}

	return commissions, total, nil
}

func (r *GormCommissionRepo) AdminList(params structs.AdminCommissionListParams) ([]database.Commission, int, error) {
	pInput := common.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := r.DB.Model(&database.Commission{})
	orm = common.Equal(orm, "status", params.Status)
	orm = common.Equal(orm, "partner_id", params.PartnerID)
	orm = common.Equal(orm, "asset", params.Asset)

	if params.Query != "" {
		q := "%" + params.Query + "%"
		orm = orm.
			Joins("LEFT JOIN users AS referred ON referred.id = commissions.referred_user_id").
			Joins("LEFT JOIN partners ON partners.id = commissions.partner_id").
			Joins("LEFT JOIN users AS partner_user ON partner_user.id = partners.user_id").
			Where("referred.email ILIKE ? OR referred.sub_account_id ILIKE ? OR commissions.referred_user_id ILIKE ? OR partner_user.email ILIKE ?", q, q, q, q)
	}

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var commissions []database.Commission
	if err := orm.
		Preload("Partner.User").
		Preload("ReferredUser").
		Order("created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&commissions).Error; err != nil {
		return nil, 0, err
	}

	return commissions, total, nil
}

func (r *GormCommissionRepo) Breakdown(partnerID string, params structs.CommissionBreakdownParams) (float64, error) {
	orm := r.DB.Model(&database.Commission{}).Where("partner_id = ?", partnerID)

	if params.StartDate != nil {
		orm = orm.Where("trade_date >= ?", params.StartDate)
	}
	if params.EndDate != nil {
		orm = orm.Where("trade_date <= ?", params.EndDate)
	}

	var sum float64
	if err := orm.Select("COALESCE(SUM(rebate_amount), 0)").Scan(&sum).Error; err != nil {
		return 0, err
	}

	return sum, nil
}

func (r *GormCommissionRepo) DailySummary(partnerID string, params structs.ChartParams) ([]dto.DailyItem, error) {
	orm := r.DB.Model(&database.Commission{}).Where("partner_id = ?", partnerID)

	if params.StartDate != nil {
		orm = orm.Where("trade_date >= ?", params.StartDate)
	}
	if params.EndDate != nil {
		orm = orm.Where("trade_date <= ?", params.EndDate)
	}

	var items []dto.DailyItem
	if err := orm.
		Select("DATE(trade_date) as date, SUM(rebate_amount) as rebate_amount, SUM(volume_usd) as trade_volume, COUNT(*) as count").
		Group("DATE(trade_date)").
		Order("date asc").
		Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}
