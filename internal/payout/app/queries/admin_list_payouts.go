package queries

import (
	"xmeta-partner/internal/payout/port"
	"xmeta-partner/structs"
)

type AdminListPayoutsHandler struct {
	Payouts port.PayoutRepo
}

func (h *AdminListPayoutsHandler) Handle(params structs.PayoutListParams) (structs.PaginationResponse, error) {
	items, total, err := h.Payouts.AdminList(params)
	if err != nil {
		return structs.PaginationResponse{}, err
	}

	return structs.PaginationResponse{Total: total, Items: items}, nil
}
