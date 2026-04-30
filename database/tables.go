package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type (
	Base struct {
		ID        string         `gorm:"primaryKey;" json:"id"`
		CreatedAt time.Time      `gorm:"column:created_at;not null" json:"createdAt"`
		UpdatedAt time.Time      `gorm:"column:updated_at;not null" json:"updatedAt"`
		DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	}
)

func (b *Base) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return
}

type (
	JSONB map[string]interface{}
)
