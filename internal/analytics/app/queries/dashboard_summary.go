package queries

import (
	"xmeta-partner/internal/analytics/app/dto"
	"xmeta-partner/internal/analytics/port"
	"xmeta-partner/structs"
)

type DashboardSummaryHandler struct {
	Dashboard port.DashboardRepo
}

func (h *DashboardSummaryHandler) Handle(partnerID string, params structs.DashboardSummaryParams) (dto.DashboardSummary, error) {
	return h.Dashboard.GetSummary(partnerID, params)
}
