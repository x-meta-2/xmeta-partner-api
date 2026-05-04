package queries

import (
	"xmeta-partner/internal/analytics/app/dto"
	"xmeta-partner/internal/analytics/port"
	"xmeta-partner/structs"
)

type CommissionTrendHandler struct {
	Analytics port.AdminAnalyticsRepo
}

func (h *CommissionTrendHandler) Handle(params structs.ChartParams) ([]dto.TrendItem, error) {
	return h.Analytics.CommissionTrend(params)
}
