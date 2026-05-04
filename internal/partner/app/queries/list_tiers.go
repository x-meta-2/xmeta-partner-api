package queries

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/partner/port"
)

type ListTiersHandler struct {
	Tiers port.TierRepo
}

func (h *ListTiersHandler) Handle() ([]database.PartnerTier, error) {
	return h.Tiers.List()
}
