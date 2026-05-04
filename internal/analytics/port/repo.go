package port

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/analytics/app/dto"
	"xmeta-partner/structs"
)

type AdminAnalyticsRepo interface {
	Summary(params structs.DashboardSummaryParams) (dto.AdminSummary, error)
	CommissionTrend(params structs.ChartParams) ([]dto.TrendItem, error)
	TopPartners(params structs.PaginationInput) ([]database.Partner, error)
	ReferralFunnel(params structs.DashboardSummaryParams) (dto.FunnelResult, error)
}

type DashboardRepo interface {
	GetSummary(partnerID string, params structs.DashboardSummaryParams) (dto.DashboardSummary, error)
	EarningsChart(partnerID string, params structs.ChartParams) ([]dto.EarningsChartItem, error)
	ReferralChart(partnerID string, params structs.ChartParams) ([]dto.ReferralChartItem, error)
}
