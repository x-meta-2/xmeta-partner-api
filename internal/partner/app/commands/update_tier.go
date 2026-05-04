package commands

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/partner/port"
	"xmeta-partner/structs"
)

type UpdateTierHandler struct {
	Tiers port.TierRepo
}

func (h *UpdateTierHandler) Handle(id string, params structs.TierUpdateParams) (database.PartnerTier, error) {
	if _, err := h.Tiers.FindByID(id); err != nil {
		return database.PartnerTier{}, err
	}

	updates := map[string]interface{}{}
	if params.Name != nil {
		updates["name"] = *params.Name
	}
	if params.Level != nil {
		updates["level"] = *params.Level
	}
	if params.CommissionRate != nil {
		updates["commission_rate"] = *params.CommissionRate
	}
	if params.MinActiveClients != nil {
		updates["min_active_clients"] = *params.MinActiveClients
	}
	if params.MinVolume != nil {
		updates["min_volume"] = *params.MinVolume
	}
	if params.MaxVolume != nil {
		updates["max_volume"] = *params.MaxVolume
	}
	if params.IsDefault != nil {
		if *params.IsDefault {
			h.Tiers.ClearDefaultExcept(id)
		}
		updates["is_default"] = *params.IsDefault
	}
	if params.Color != nil {
		updates["color"] = *params.Color
	}

	if len(updates) > 0 {
		if err := h.Tiers.Update(id, updates); err != nil {
			return database.PartnerTier{}, err
		}
	}

	tier, err := h.Tiers.FindByID(id)
	if err != nil {
		return database.PartnerTier{}, err
	}
	return *tier, nil
}
