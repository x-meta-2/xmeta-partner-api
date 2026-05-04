package queries

import (
	"xmeta-partner/internal/commission/port"
	"xmeta-partner/structs"
)

type ListCommissionsHandler struct {
	Commissions port.CommissionRepo
}

func (h *ListCommissionsHandler) Handle(partnerID string, params structs.CommissionListParams) (structs.PaginationResponse, error) {
	items, total, err := h.Commissions.List(partnerID, params)
	if err != nil {
		return structs.PaginationResponse{}, err
	}

	return structs.PaginationResponse{Total: total, Items: items}, nil
}
