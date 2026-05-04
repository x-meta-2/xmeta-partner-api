package queries

import (
	"xmeta-partner/internal/analytics/app/dto"
	"xmeta-partner/internal/analytics/port"
	"xmeta-partner/structs"
)

type ReferralFunnelHandler struct {
	Analytics port.AdminAnalyticsRepo
}

func (h *ReferralFunnelHandler) Handle(params structs.DashboardSummaryParams) (dto.FunnelResult, error) {
	return h.Analytics.ReferralFunnel(params)
}
