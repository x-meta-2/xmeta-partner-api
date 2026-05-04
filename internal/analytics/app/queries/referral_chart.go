package queries

import (
	"xmeta-partner/internal/analytics/app/dto"
	"xmeta-partner/internal/analytics/port"
	"xmeta-partner/structs"
)

type ReferralChartHandler struct {
	Dashboard port.DashboardRepo
}

func (h *ReferralChartHandler) Handle(partnerID string, params structs.ChartParams) ([]dto.ReferralChartItem, error) {
	return h.Dashboard.ReferralChart(partnerID, params)
}
