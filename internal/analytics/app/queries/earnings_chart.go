package queries

import (
	"xmeta-partner/internal/analytics/app/dto"
	"xmeta-partner/internal/analytics/port"
	"xmeta-partner/structs"
)

type EarningsChartHandler struct {
	Dashboard port.DashboardRepo
}

func (h *EarningsChartHandler) Handle(partnerID string, params structs.ChartParams) ([]dto.EarningsChartItem, error) {
	return h.Dashboard.EarningsChart(partnerID, params)
}
