package admin

import (
	"xmeta-partner/database"
	"xmeta-partner/services"
	"xmeta-partner/structs"
)

type AnalyticsService struct {
	services.BaseService
}

// Summary returns overall partner program analytics:
// total partners, active partners, pending applications, total commissions paid, total volume
func (s *AnalyticsService) Summary(params structs.DashboardSummaryParams) (map[string]interface{}, error) {
	var totalPartners int64
	s.DB.Model(&database.Partner{}).Count(&totalPartners)

	var activePartners int64
	s.DB.Model(&database.Partner{}).Where("status = ?", "active").Count(&activePartners)

	var pendingApplications int64
	s.DB.Model(&database.PartnerApplication{}).Where("status = ?", "pending").Count(&pendingApplications)

	var totalCommissions float64
	s.DB.Model(&database.Commission{}).
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&totalCommissions)

	var totalVolume float64
	s.DB.Model(&database.Commission{}).
		Select("COALESCE(SUM(trade_amount), 0)").
		Scan(&totalVolume)

	var totalReferrals int64
	s.DB.Model(&database.Referral{}).Count(&totalReferrals)

	var pendingPayouts float64
	s.DB.Model(&database.Payout{}).
		Where("status = ?", "pending").
		Select("COALESCE(SUM(amount), 0)").
		Scan(&pendingPayouts)

	return map[string]interface{}{
		"totalPartners":       totalPartners,
		"activePartners":      activePartners,
		"pendingApplications": pendingApplications,
		"totalCommissions":    totalCommissions,
		"totalVolume":         totalVolume,
		"totalReferrals":      totalReferrals,
		"pendingPayouts":      pendingPayouts,
	}, nil
}

func (s *AnalyticsService) CommissionTrend(params structs.ChartParams) (interface{}, error) {
	type TrendItem struct {
		Date             string  `json:"date"`
		TotalCommission  float64 `json:"totalCommission"`
		TotalVolume      float64 `json:"totalVolume"`
		TransactionCount int64   `json:"transactionCount"`
	}

	var items []TrendItem
	orm := s.DB.Model(&database.Commission{}).
		Select("DATE(trade_date) as date, SUM(commission_amount) as total_commission, SUM(trade_amount) as total_volume, COUNT(*) as transaction_count").
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

func (s *AnalyticsService) TopPartners(params structs.PaginationInput) (interface{}, error) {
	pInput := services.PreparePaginationInput(params)
	if pInput.PageSize > 10 {
		pInput.PageSize = 10
	}

	var partners []database.Partner
	if err := s.DB.Preload("Tier").
		Where("status = ?", "active").
		Order("total_earnings desc").
		Offset((pInput.Current - 1) * pInput.PageSize).
		Limit(pInput.PageSize).
		Find(&partners).Error; err != nil {
		return nil, err
	}

	return partners, nil
}

func (s *AnalyticsService) ReferralFunnel(params structs.DashboardSummaryParams) (map[string]interface{}, error) {
	var totalClicks int64
	s.DB.Model(&database.ReferralLink{}).Select("COALESCE(SUM(clicks), 0)").Scan(&totalClicks)

	var totalRegistrations int64
	s.DB.Model(&database.Referral{}).Count(&totalRegistrations)

	var totalDeposits int64
	s.DB.Model(&database.Referral{}).Where("first_deposit_at IS NOT NULL").Count(&totalDeposits)

	var totalTraders int64
	s.DB.Model(&database.Referral{}).Where("first_trade_at IS NOT NULL").Count(&totalTraders)

	return map[string]interface{}{
		"clicks":        totalClicks,
		"registrations": totalRegistrations,
		"deposits":      totalDeposits,
		"traders":       totalTraders,
	}, nil
}
