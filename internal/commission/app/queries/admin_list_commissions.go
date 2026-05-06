package queries

import (
	"xmeta-partner/internal/commission/port"
	"xmeta-partner/structs"
)

type AdminListCommissionsHandler struct {
	Commissions port.CommissionRepo
}

func (h *AdminListCommissionsHandler) Handle(params structs.AdminCommissionListParams) (structs.PaginationResponse, error) {
	items, total, err := h.Commissions.AdminList(params)
	if err != nil {
		return structs.PaginationResponse{}, err
	}

	return structs.PaginationResponse{Total: total, Items: items}, nil
}
