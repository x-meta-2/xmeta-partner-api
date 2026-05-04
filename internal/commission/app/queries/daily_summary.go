package queries

import (
	"xmeta-partner/internal/commission/app/dto"
	"xmeta-partner/internal/commission/port"
	"xmeta-partner/structs"
)

type DailySummaryHandler struct {
	Commissions port.CommissionRepo
}

func (h *DailySummaryHandler) Handle(partnerID string, params structs.ChartParams) ([]dto.DailyItem, error) {
	return h.Commissions.DailySummary(partnerID, params)
}
