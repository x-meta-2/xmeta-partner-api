package analytics

import (
	"xmeta-partner/internal/analytics/adapters"
	"xmeta-partner/internal/analytics/app/queries"

	"gorm.io/gorm"
)

type Service struct {
	Queries Queries
}

type Queries struct {
	AdminSummary    *queries.AdminSummaryHandler
	CommissionTrend *queries.CommissionTrendHandler
	TopPartners     *queries.TopPartnersHandler
	ReferralFunnel  *queries.ReferralFunnelHandler

	DashboardSummary *queries.DashboardSummaryHandler
	EarningsChart    *queries.EarningsChartHandler
	ReferralChart    *queries.ReferralChartHandler
}

func NewService(db *gorm.DB) *Service {
	adminRepo := &adapters.GormAdminAnalyticsRepo{DB: db}
	dashRepo := &adapters.GormDashboardRepo{DB: db}

	return &Service{
		Queries: Queries{
			AdminSummary:     &queries.AdminSummaryHandler{Analytics: adminRepo},
			CommissionTrend:  &queries.CommissionTrendHandler{Analytics: adminRepo},
			TopPartners:      &queries.TopPartnersHandler{Analytics: adminRepo},
			ReferralFunnel:   &queries.ReferralFunnelHandler{Analytics: adminRepo},
			DashboardSummary: &queries.DashboardSummaryHandler{Dashboard: dashRepo},
			EarningsChart:    &queries.EarningsChartHandler{Dashboard: dashRepo},
			ReferralChart:    &queries.ReferralChartHandler{Dashboard: dashRepo},
		},
	}
}
