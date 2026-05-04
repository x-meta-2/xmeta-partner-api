package queries

import (
	"xmeta-partner/internal/commission/app/dto"
	"xmeta-partner/internal/commission/port"
	"xmeta-partner/structs"
)

type CommissionBreakdownHandler struct {
	Commissions port.CommissionRepo
}

func (h *CommissionBreakdownHandler) Handle(partnerID string, params structs.CommissionBreakdownParams) (dto.CommissionBreakdown, error) {
	sum, err := h.Commissions.Breakdown(partnerID, params)
	if err != nil {
		return dto.CommissionBreakdown{}, err
	}

	return dto.CommissionBreakdown{Futures: sum, Total: sum}, nil
}
