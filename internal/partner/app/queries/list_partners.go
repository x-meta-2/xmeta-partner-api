package queries

import (
	"xmeta-partner/internal/partner/port"
	"xmeta-partner/structs"
)

type ListPartnersHandler struct {
	Partners port.PartnerRepo
}

func (h *ListPartnersHandler) Handle(params structs.PartnerListParams) (structs.PaginationResponse, error) {
	items, total, err := h.Partners.List(params)
	if err != nil {
		return structs.PaginationResponse{}, err
	}

	return structs.PaginationResponse{Total: total, Items: items}, nil
}
