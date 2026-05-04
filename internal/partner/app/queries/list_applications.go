package queries

import (
	"xmeta-partner/internal/partner/port"
	"xmeta-partner/structs"
)

type ListApplicationsHandler struct {
	Apps port.ApplicationRepo
}

func (h *ListApplicationsHandler) Handle(params structs.ApplicationListParams) (structs.PaginationResponse, error) {
	items, total, err := h.Apps.List(params)
	if err != nil {
		return structs.PaginationResponse{}, err
	}

	return structs.PaginationResponse{Total: total, Items: items}, nil
}
