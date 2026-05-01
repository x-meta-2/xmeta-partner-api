package partner

import (
	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/services"
	"xmeta-partner/structs"
)

type CommissionService struct {
	services.BaseService
}

func (s *CommissionService) List(partnerID string, params structs.CommissionListParams) (structs.PaginationResponse, error) {
	pInput := services.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := s.DB.Model(&database.Commission{}).Where("partner_id = ?", partnerID)
	orm = common.Equal(orm, "status", params.Status)
	orm = common.Equal(orm, "referred_user_id", params.ReferredUserID)

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var commissions []database.Commission
	if err := orm.
		Order("created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&commissions).Error; err != nil {
		return structs.PaginationResponse{}, err
	}

	return structs.PaginationResponse{Total: total, Items: commissions}, nil
}

// Breakdown returns the partner's total commission earnings within an
// optional date window. Futures is the only trade type today, so there is
// no per-product split — the field is kept as a single total.
type CommissionBreakdown struct {
	Futures float64 `json:"futures"`
	Total   float64 `json:"total"`
}

func (s *CommissionService) Breakdown(partnerID string, params structs.CommissionBreakdownParams) (CommissionBreakdown, error) {
	orm := s.DB.Model(&database.Commission{}).
		Where("partner_id = ?", partnerID).
		Select("COALESCE(SUM(commission_amount), 0) AS sum")

	if params.StartDate != nil {
		orm = orm.Where("trade_date >= ?", params.StartDate)
	}
	if params.EndDate != nil {
		orm = orm.Where("trade_date <= ?", params.EndDate)
	}

	var sum float64
	if err := orm.Scan(&sum).Error; err != nil {
		return CommissionBreakdown{}, err
	}

	return CommissionBreakdown{Futures: sum, Total: sum}, nil
}

func (s *CommissionService) DailySummary(partnerID string, params structs.ChartParams) (interface{}, error) {
	type DailyItem struct {
		Date             string  `json:"date"`
		CommissionAmount float64 `json:"commissionAmount"`
		TradeVolume      float64 `json:"tradeVolume"`
		Count            int64   `json:"count"`
	}

	var items []DailyItem
	orm := s.DB.Model(&database.Commission{}).
		Where("partner_id = ?", partnerID).
		Select("DATE(trade_date) as date, SUM(commission_amount) as commission_amount, SUM(trade_amount) as trade_volume, COUNT(*) as count").
		Group("DATE(trade_date)").
		Order("date asc")

	if params.StartDate != nil {
		orm = orm.Where("trade_date >= ?", params.StartDate)
	}
	if params.EndDate != nil {
		orm = orm.Where("trade_date <= ?", params.EndDate)
	}

	if err := orm.Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}
