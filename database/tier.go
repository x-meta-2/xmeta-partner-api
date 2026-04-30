package database

// PartnerTier — commission rate tier (Standard, Bronze, Silver, etc.).
//
// Commission is futures-only — `commission_rate` is the futures rate for
// this tier. There is no per-trade-type override layer.
type PartnerTier struct {
	Base
	Name             string   `gorm:"column:name;not null;unique" json:"name"`
	Level            int      `gorm:"column:level;not null;unique" json:"level"`
	CommissionRate   float64  `gorm:"column:commission_rate;type:decimal(5,4);not null" json:"commissionRate"`
	MinActiveClients int      `gorm:"column:min_active_clients;default:0" json:"minActiveClients"`
	MinVolume        float64  `gorm:"column:min_volume;type:decimal(20,8);default:0" json:"minVolume"`
	MaxVolume        *float64 `gorm:"column:max_volume;type:decimal(20,8)" json:"maxVolume"`
	IsDefault        bool     `gorm:"column:is_default;default:false" json:"isDefault"`
	Color            string   `gorm:"column:color" json:"color"`
}
