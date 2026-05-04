package adapters

import (
	"xmeta-partner/database"

	"gorm.io/gorm"
)

type GormPayoutItemRepo struct {
	DB *gorm.DB
}

func (r *GormPayoutItemRepo) CreateBatch(items []database.PayoutItem) error {
	for i := range items {
		if err := r.DB.Create(&items[i]).Error; err != nil {
			return err
		}
	}
	return nil
}
