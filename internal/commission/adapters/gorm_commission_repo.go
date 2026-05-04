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

func (r *GormCommissionRepo) Breakdown(partnerID string, params structs.CommissionBreakdownParams) (float64, error) {
	orm := r.DB.Model(&database.Commission{}).Where("partner_id = ?", partnerID)

	if params.StartDate != nil {
		orm = orm.Where("trade_date >= ?", params.StartDate)
	}
	if params.EndDate != nil {
		orm = orm.Where("trade_date <= ?", params.EndDate)
	}

	var sum float64
	if err := orm.Select("COALESCE(SUM(commission_amount), 0)").Scan(&sum).Error; err != nil {
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
		Select("DATE(trade_date) as date, SUM(commission_amount) as commission_amount, SUM(trade_amount) as trade_volume, COUNT(*) as count").
		Group("DATE(trade_date)").
		Order("date asc").
		Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}
