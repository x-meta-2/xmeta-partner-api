package adapters

import (
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
	result.TotalReferrals = partner.TotalReferrals

	r.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND status = ?", partnerID, "pending").
		Select("COALESCE(SUM(rebate_amount), 0)").
		Scan(&result.PendingCommission)

	r.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND DATE_TRUNC('month', created_at) = DATE_TRUNC('month', NOW())", partnerID).
		Select("COALESCE(SUM(rebate_amount), 0)").
		Scan(&result.MonthEarnings)

	r.DB.Model(&database.Referral{}).
		Where("partner_id = ? AND ended_at IS NULL", partnerID).
		Count(&result.ActiveReferrals)

	r.DB.Model(&database.Commission{}).
		Where("partner_id = ?", partnerID).
		Select("COALESCE(SUM(volume_usd), 0)").
		Scan(&result.TotalVolume)

	var tradingReferrals int64
	r.DB.Model(&database.Referral{}).
		Where("partner_id = ? AND ended_at IS NULL AND first_trade_at IS NOT NULL", partnerID).
		Count(&tradingReferrals)

	if result.ActiveReferrals > 0 {
		result.ConversionRate = float64(tradingReferrals) / float64(result.ActiveReferrals) * 100
	}

	return result, nil
}

func (r *GormDashboardRepo) EarningsChart(partnerID string, params structs.ChartParams) ([]dto.EarningsChartItem, error) {
	var items []dto.EarningsChartItem

	orm := r.DB.Model(&database.Commission{}).
		Select("DATE(trade_date) as date, SUM(rebate_amount) as commissions, SUM(volume_usd) as trade_volume").
		Where("partner_id = ?", partnerID)

	if params.StartDate != nil {
		orm = orm.Where("trade_date >= ?", params.StartDate)
	}
	if params.EndDate != nil {
		orm = orm.Where("trade_date <= ?", params.EndDate)
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

	if params.StartDate != nil {
		orm = orm.Where("registered_at >= ?", params.StartDate)
	}
	if params.EndDate != nil {
		orm = orm.Where("registered_at <= ?", params.EndDate)
	}

	if err := orm.Group("DATE(registered_at)").Order("date asc").Find(&items).Error; err != nil {
		return nil, err
	}

	return items, nil
}
