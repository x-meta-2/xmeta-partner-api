package queries

import (
	"xmeta-partner/internal/referral/port"
	"xmeta-partner/structs"
)

type ListLinksHandler struct {
	Links port.ReferralLinkRepo
}

func (h *ListLinksHandler) Handle(partnerID string, params structs.ReferralListParams) (structs.PaginationResponse, error) {
	links, total, err := h.Links.List(partnerID, params)
	if err != nil {
		return structs.PaginationResponse{}, err
	}
	return structs.PaginationResponse{Total: total, Items: links}, nil
}
