package commands

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/partner/port"
)

type UpdatePartnerStatusHandler struct {
	Partners port.PartnerRepo
}

func (h *UpdatePartnerStatusHandler) Handle(partnerID string, status string) (*database.Partner, error) {
	if err := h.Partners.UpdateStatus(partnerID, status); err != nil {
		return nil, err
	}

	return h.Partners.FindByID(partnerID)
}
