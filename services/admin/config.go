package admin

import (
	"fmt"

	"xmeta-partner/database"
	"xmeta-partner/services"
	"xmeta-partner/structs"
)

type ConfigService struct {
	services.BaseService
}

func (s *ConfigService) ListTiers() ([]database.PartnerTier, error) {
	var tiers []database.PartnerTier
	if err := s.DB.Order("level asc").Find(&tiers).Error; err != nil {
		return nil, err
	}
	return tiers, nil
}

func (s *ConfigService) CreateTier(params structs.TierCreateParams) (database.PartnerTier, error) {
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
		s.DB.Model(&database.PartnerTier{}).Where("is_default = ?", true).Update("is_default", false)
	}

	if err := s.DB.Create(&tier).Error; err != nil {
		return database.PartnerTier{}, err
	}

	return tier, nil
}

func (s *ConfigService) UpdateTier(id string, params structs.TierUpdateParams) (database.PartnerTier, error) {
	var tier database.PartnerTier
	if err := s.DB.Where("id = ?", id).First(&tier).Error; err != nil {
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
			s.DB.Model(&database.PartnerTier{}).Where("is_default = ?", true).Update("is_default", false)
		}
		updates["is_default"] = *params.IsDefault
	}
	if params.Color != nil {
		updates["color"] = *params.Color
	}

	if len(updates) > 0 {
		if err := s.DB.Model(&tier).Updates(updates).Error; err != nil {
			return database.PartnerTier{}, err
		}
	}

	s.DB.Where("id = ?", id).First(&tier)
	return tier, nil
}

// DeleteTier soft deletes a tier (only if no partners use it)
func (s *ConfigService) DeleteTier(id string) error {
	var count int64
	s.DB.Model(&database.Partner{}).Where("tier_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("cannot delete tier: %d partners are using it", count)
	}

	result := s.DB.Where("id = ?", id).Delete(&database.PartnerTier{})
	if result.RowsAffected == 0 {
		return fmt.Errorf("tier not found")
	}
	return result.Error
}
