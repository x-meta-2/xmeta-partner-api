package commands

import (
	"fmt"

	"xmeta-partner/internal/partner/port"
)

type DeleteTierHandler struct {
	Tiers port.TierRepo
}

func (h *DeleteTierHandler) Handle(id string) error {
	count, err := h.Tiers.CountPartnersByTier(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("cannot delete tier: %d partners are using it", count)
	}

	return h.Tiers.Delete(id)
}
