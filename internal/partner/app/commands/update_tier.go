package commands

import (
	"fmt"

	"xmeta-partner/database"
	"xmeta-partner/internal/partner/port"
	"xmeta-partner/structs"
)

type UpdateTierHandler struct {
	Tiers port.TierRepo
}

func (h *UpdateTierHandler) Handle(id string, params structs.TierUpdateParams) (database.PartnerTier, error) {
	current, err := h.Tiers.FindByID(id)
	if err != nil {
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
		if !*params.IsDefault && current.IsDefault {
			return database.PartnerTier{}, fmt.Errorf("cannot unset the default tier")
		}
		if *params.IsDefault {
			updates["is_default"] = true
		} else {
			updates["is_default"] = false
		}
	}
	if params.Color != nil {
		updates["color"] = *params.Color
	}

	if len(updates) > 0 {
		err := func() error {
			if params.IsDefault != nil && *params.IsDefault {
				return h.Tiers.RunInTx(func(txTiers port.TierRepo) error {
					if err := txTiers.ClearDefaultExcept(id); err != nil {
						return err
					}
					return txTiers.Update(id, updates)
				})
			}
			return h.Tiers.Update(id, updates)
		}()
		if err != nil {
			return database.PartnerTier{}, err
		}
	}

	tier, err := h.Tiers.FindByID(id)
	if err != nil {
		return database.PartnerTier{}, err
	}
	return *tier, nil
}
