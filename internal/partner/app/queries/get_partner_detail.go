package queries

import (
	"xmeta-partner/internal/partner/app/dto"
	"xmeta-partner/internal/partner/port"
)

type GetPartnerDetailHandler struct {
	Detail port.PartnerDetailRepo
}

func (h *GetPartnerDetailHandler) Handle(id string) (*dto.PartnerDetail, error) {
	return h.Detail.GetDetail(id)
}
