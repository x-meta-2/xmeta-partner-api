package commands

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/partner/port"
	"xmeta-partner/structs"
)

type CreateTierHandler struct {
	Tiers port.TierRepo
}

func (h *CreateTierHandler) Handle(params structs.TierCreateParams) (database.PartnerTier, error) {
	tier := database.PartnerTier{
		Name:             params.Name,
		Level:            params.Level,
		CommissionRate:   params.CommissionRate,
		MinActiveClients: params.MinActiveClients,
		MinVolume:        params.MinVolume,
		MaxVolume:        params.MaxVolume,
		IsDefault:        params.IsDefault,
		Color:            params.Color,
	}

	if params.IsDefault {
		if err := h.Tiers.RunInTx(func(txTiers port.TierRepo) error {
			if err := txTiers.ClearDefaultExcept(""); err != nil {
				return err
			}
			return txTiers.Create(&tier)
		}); err != nil {
			return database.PartnerTier{}, err
		}
		return tier, nil
	}

	if err := h.Tiers.Create(&tier); err != nil {
		return database.PartnerTier{}, err
	}

	return tier, nil
}
