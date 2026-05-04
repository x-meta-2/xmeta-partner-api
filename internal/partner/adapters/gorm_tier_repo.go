package adapters

import (
	"xmeta-partner/database"

	"gorm.io/gorm"
)

type GormTierRepo struct {
	DB *gorm.DB
}

func (r *GormTierRepo) FindByID(id string) (*database.PartnerTier, error) {
	var tier database.PartnerTier
	if err := r.DB.Where("id = ?", id).First(&tier).Error; err != nil {
		return nil, err
	}
	return &tier, nil
}

func (r *GormTierRepo) FindDefault() (*database.PartnerTier, error) {
	var tier database.PartnerTier
	if err := r.DB.Where("is_default = ?", true).First(&tier).Error; err != nil {
		return nil, err
	}
	return &tier, nil
}

func (r *GormTierRepo) List() ([]database.PartnerTier, error) {
	var tiers []database.PartnerTier
	if err := r.DB.Order("level asc").Find(&tiers).Error; err != nil {
		return nil, err
	}
	return tiers, nil
}

func (r *GormTierRepo) Create(tier *database.PartnerTier) error {
	return r.DB.Create(tier).Error
}

func (r *GormTierRepo) Update(id string, fields map[string]interface{}) error {
	return r.DB.Model(&database.PartnerTier{}).Where("id = ?", id).Updates(fields).Error
}

func (r *GormTierRepo) Delete(id string) error {
	return r.DB.Where("id = ?", id).Delete(&database.PartnerTier{}).Error
}

func (r *GormTierRepo) CountPartnersByTier(tierID string) (int64, error) {
	var count int64
	if err := r.DB.Model(&database.Partner{}).Where("tier_id = ?", tierID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *GormTierRepo) ClearDefaultExcept(tierID string) error {
	return r.DB.Model(&database.PartnerTier{}).
		Where("id != ? AND is_default = ?", tierID, true).
		Update("is_default", false).Error
}
