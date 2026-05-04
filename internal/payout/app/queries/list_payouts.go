package queries

import (
	"xmeta-partner/internal/payout/port"
	"xmeta-partner/structs"
)

type ListPayoutsHandler struct {
	Payouts port.PayoutRepo
}

func (h *ListPayoutsHandler) Handle(partnerID string, params structs.PayoutListParams) (structs.PaginationResponse, error) {
	items, total, err := h.Payouts.List(partnerID, params)
	if err != nil {
		return structs.PaginationResponse{}, err
	}

	return structs.PaginationResponse{Total: total, Items: items}, nil
}
