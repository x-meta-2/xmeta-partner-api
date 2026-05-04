package commands

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/partner/domain"
	"xmeta-partner/internal/partner/port"
)

type UpdatePartnerTierHandler struct {
	Partners port.PartnerRepo
	Tiers    port.TierRepo
}

func (h *UpdatePartnerTierHandler) Handle(partnerID string, tierID string) (*database.Partner, error) {
	if _, err := h.Tiers.FindByID(tierID); err != nil {
		return nil, domain.ErrTierNotFound
	}

	if err := h.Partners.UpdateTier(partnerID, tierID); err != nil {
		return nil, err
	}

	return h.Partners.FindByID(partnerID)
}
