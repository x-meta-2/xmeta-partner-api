package queries

import (
	"xmeta-partner/internal/analytics/app/dto"
	"xmeta-partner/internal/analytics/port"
	"xmeta-partner/structs"
)

type AdminSummaryHandler struct {
	Analytics port.AdminAnalyticsRepo
}

func (h *AdminSummaryHandler) Handle(params structs.DashboardSummaryParams) (dto.AdminSummary, error) {
	return h.Analytics.Summary(params)
}
