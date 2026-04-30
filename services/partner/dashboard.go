package partner

import (
	"xmeta-partner/database"
	"xmeta-partner/services"
	"xmeta-partner/structs"
)

type DashboardService struct {
	services.BaseService
}

func (s *DashboardService) GetSummary(partnerID string, params structs.DashboardSummaryParams) (map[string]interface{}, error) {
	var partner database.Partner
	if err := s.DB.Where("id = ?", partnerID).First(&partner).Error; err != nil {
		return nil, err
	}

	var pendingAmount float64
	s.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND status = ?", partnerID, "pending").
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&pendingAmount)

	var monthEarnings float64
	s.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())", partnerID).
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&monthEarnings)

	// "Active referrals" = users currently linked to this partner. Past
	// rows (ended_at NOT NULL — user switched away) stay in the table for
	// commission attribution but must not inflate the live count.
	var activeReferrals int64
	s.DB.Model(&database.Referral{}).
		Where("partner_id = ? AND ended_at IS NULL", partnerID).
		Count(&activeReferrals)

	var totalVolume float64
	s.DB.Model(&database.Commission{}).
		Where("partner_id = ?", partnerID).
		Select("COALESCE(SUM(trade_amount), 0)").
		Scan(&totalVolume)

	// Conversion = currently-linked users who have traded ÷ currently-linked
	// users. Both numerator and denominator filter `ended_at IS NULL` so
	// historical relationships don't skew the rate.
	var tradingReferrals int64
	s.DB.Model(&database.Referral{}).
		Where("partner_id = ? AND ended_at IS NULL AND first_trade_at IS NOT NULL", partnerID).
		Count(&tradingReferrals)

	conversionRate := float64(0)
	if activeReferrals > 0 {
		conversionRate = float64(tradingReferrals) / float64(activeReferrals) * 100
	}

	return map[string]interface{}{
		"totalEarnings":     partner.TotalEarnings,
		"monthEarnings":     monthEarnings,
		"pendingCommission": pendingAmount,
		"totalReferrals":    partner.TotalReferrals,
		"activeReferrals":   activeReferrals,
		"totalVolume":       totalVolume,
		"conversionRate":    conversionRate,
	}, nil
}

func (s *DashboardService) GetEarningsChart(partnerID string, params structs.ChartParams) (interface{}, error) {
	type ChartItem struct {
		Date        string  `json:"date"`
		Commissions float64 `json:"commissions"`
		TradeVolume float64 `json:"tradeVolume"`
	}

	var items []ChartItem
	orm := s.DB.Model(&database.Commission{}).
		Where("partner_id = ?", partnerID).
		Select("DATE(trade_date) as date, SUM(commission_amount) as commissions, SUM(trade_amount) as trade_volume").
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

func (s *DashboardService) GetReferralChart(partnerID string, params structs.ChartParams) (interface{}, error) {
	type ChartItem struct {
		Date    string `json:"date"`
		Signups int64  `json:"signups"`
	}

	var items []ChartItem
	orm := s.DB.Model(&database.Referral{}).
		Where("partner_id = ?", partnerID).
		Select("DATE(registered_at) as date, COUNT(*) as signups").
		Group("DATE(registered_at)").
		Order("date asc")

	if params.StartDate != nil {
		orm = orm.Where("registered_at >= ?", params.StartDate)
	}
	if params.EndDate != nil {
		orm = orm.Where("registered_at <= ?", params.EndDate)
	}

	if err := orm.Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (s *DashboardService) GetTierProgress(partnerID string) (interface{}, error) {
	authService := AuthService{BaseService: s.BaseService}
	return authService.GetTierDetails(partnerID)
}
