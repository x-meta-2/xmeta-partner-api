package adapters

import (
	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/internal/analytics/app/dto"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type GormAdminAnalyticsRepo struct {
	DB *gorm.DB
}

func (r *GormAdminAnalyticsRepo) Summary(params structs.DashboardSummaryParams) (dto.AdminSummary, error) {
	var result dto.AdminSummary

	if err := r.DB.Model(&database.Partner{}).Count(&result.TotalPartners).Error; err != nil {
		return result, err
	}

	if err := r.DB.Model(&database.Partner{}).Where("status = ?", "active").Count(&result.ActivePartners).Error; err != nil {
		return result, err
	}

	if err := r.DB.Model(&database.PartnerApplication{}).Where("status = ?", "pending").Count(&result.PendingApplications).Error; err != nil {
		return result, err
	}

	if err := r.DB.Model(&database.Commission{}).Select("COALESCE(SUM(commission_amount), 0)").Scan(&result.TotalCommissions).Error; err != nil {
		return result, err
	}

	if err := r.DB.Model(&database.Commission{}).Select("COALESCE(SUM(trade_amount), 0)").Scan(&result.TotalVolume).Error; err != nil {
		return result, err
	}

	if err := r.DB.Model(&database.Referral{}).Count(&result.TotalReferrals).Error; err != nil {
		return result, err
	}

	if err := r.DB.Model(&database.Payout{}).Where("status = ?", "pending").Select("COALESCE(SUM(amount), 0)").Scan(&result.PendingPayouts).Error; err != nil {
		return result, err
	}

	return result, nil
}

func (r *GormAdminAnalyticsRepo) CommissionTrend(params structs.ChartParams) ([]dto.TrendItem, error) {
	var items []dto.TrendItem

	orm := r.DB.Model(&database.Commission{}).
		Select("DATE(trade_date) as date, SUM(commission_amount) as total_commission, SUM(trade_amount) as total_volume, COUNT(*) as transaction_count")

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

func (r *GormAdminAnalyticsRepo) TopPartners(params structs.PaginationInput) ([]database.Partner, error) {
	pInput := common.PreparePaginationInput(params)

	if pInput.PageSize > 10 {
		pInput.PageSize = 10
	}

	offset := (pInput.Current - 1) * pInput.PageSize

	var partners []database.Partner
	if err := r.DB.Preload("Tier").
		Where("status = ?", "active").
		Order("total_earnings desc").
		Offset(offset).
		Limit(pInput.PageSize).
		Find(&partners).Error; err != nil {
		return nil, err
	}

	return partners, nil
}

func (r *GormAdminAnalyticsRepo) ReferralFunnel(params structs.DashboardSummaryParams) (dto.FunnelResult, error) {
	var result dto.FunnelResult

	if err := r.DB.Model(&database.ReferralLink{}).Select("COALESCE(SUM(clicks), 0)").Scan(&result.Clicks).Error; err != nil {
		return result, err
	}

	if err := r.DB.Model(&database.Referral{}).Count(&result.Registrations).Error; err != nil {
		return result, err
	}

	if err := r.DB.Model(&database.Referral{}).Where("first_trade_at IS NOT NULL").Count(&result.Traders).Error; err != nil {
		return result, err
	}

	return result, nil
}
