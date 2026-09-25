package adapters

import (
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/analytics/app/dto"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type GormDashboardRepo struct {
	DB *gorm.DB
}

func (r *GormDashboardRepo) GetSummary(partnerID string, params structs.DashboardSummaryParams) (dto.DashboardSummary, error) {
	var result dto.DashboardSummary

	var partner database.Partner
	if err := r.DB.Where("id = ?", partnerID).First(&partner).Error; err != nil {
		return result, err
	}
	result.TotalEarnings = partner.TotalEarnings

	if err := r.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND status = ?", partnerID, database.CommissionStatusPending).
		Select("COALESCE(SUM(rebate_amount), 0)").
		Scan(&result.PendingCommission).Error; err != nil {
		return result, err
	}

	location, err := time.LoadLocation("Asia/Ulaanbaatar")
	if err != nil {
		location = time.FixedZone("UTC+8", 8*60*60)
	}
	now := time.Now().In(location)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)
	nextMonthStart := monthStart.AddDate(0, 1, 0)

	if err := r.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND trade_date >= ? AND trade_date < ?", partnerID, monthStart, nextMonthStart).
		Select("COALESCE(SUM(rebate_amount), 0)").
		Scan(&result.MonthEarnings).Error; err != nil {
		return result, err
	}

	var totalReferrals int64
	if err := r.DB.Model(&database.Referral{}).
		Where("partner_id = ? AND ended_at IS NULL", partnerID).
		Count(&totalReferrals).Error; err != nil {
		return result, err
	}
	result.TotalReferrals = int(totalReferrals)

	if err := r.DB.Model(&database.Referral{}).
		Where("partner_id = ? AND status = ? AND ended_at IS NULL", partnerID, database.ReferralStatusActive).
		Count(&result.ActiveReferrals).Error; err != nil {
		return result, err
	}

	if err := r.DB.Model(&database.Commission{}).
		Where("partner_id = ?", partnerID).
		Select("COALESCE(SUM(volume_usd), 0)").
		Scan(&result.TotalVolume).Error; err != nil {
		return result, err
	}

	if result.TotalReferrals > 0 {
		result.ConversionRate = float64(result.ActiveReferrals) / float64(result.TotalReferrals) * 100
	}

	return result, nil
}

func (r *GormDashboardRepo) EarningsChart(partnerID string, params structs.ChartParams) ([]dto.EarningsChartItem, error) {
	var items []dto.EarningsChartItem

	orm := r.DB.Model(&database.Commission{}).
		Select("DATE(trade_date) as date, SUM(rebate_amount) as commissions, SUM(volume_usd) as trade_volume").
		Where("partner_id = ?", partnerID)

	start, end := chartRange(params)
	if start != nil {
		orm = orm.Where("trade_date >= ?", start)
	}
	if end != nil {
		orm = orm.Where("trade_date < ?", end)
	}

	if err := orm.Group("DATE(trade_date)").Order("date asc").Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func (r *GormDashboardRepo) ReferralChart(partnerID string, params structs.ChartParams) ([]dto.ReferralChartItem, error) {
	var items []dto.ReferralChartItem

	orm := r.DB.Model(&database.Referral{}).
		Select("DATE(registered_at) as date, COUNT(*) as signups").
		Where("partner_id = ?", partnerID)

	start, end := chartRange(params)
	if start != nil {
		orm = orm.Where("registered_at >= ?", start)
	}
	if end != nil {
		orm = orm.Where("registered_at < ?", end)
	}

	if err := orm.Group("DATE(registered_at)").Order("date asc").Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}

func chartRange(params structs.ChartParams) (*time.Time, *time.Time) {
	if params.StartDate != nil || params.EndDate != nil {
		return params.StartDate, params.EndDate
	}

	days := map[string]int{"7d": 7, "30d": 30, "90d": 90}
	location, err := time.LoadLocation("Asia/Ulaanbaatar")
	if err != nil {
		location = time.FixedZone("UTC+8", 8*60*60)
	}

	var start time.Time
	now := time.Now().In(location)
	end := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, location)
	if periodDays, ok := days[params.Period]; ok {
		start = end.AddDate(0, 0, -periodDays)
	} else if params.Period == "1y" {
		start = end.AddDate(-1, 0, 0)
	} else {
		return nil, nil
	}

	return &start, &end
}
