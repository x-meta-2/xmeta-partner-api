package queries

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/partner/port"
)

type GetApplicationHandler struct {
	Apps port.ApplicationRepo
}

func (h *GetApplicationHandler) Handle(id string) (*database.PartnerApplication, error) {
	return h.Apps.FindByID(id)
}
