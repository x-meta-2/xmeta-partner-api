package database

type PartnerDailyStats struct {
	Base
	PartnerID       string  `gorm:"column:partner_id;not null;uniqueIndex:idx_partner_date" json:"partnerId"`
	Date            string  `gorm:"column:date;not null;uniqueIndex:idx_partner_date" json:"date"`
	Clicks          int     `gorm:"column:clicks;default:0" json:"clicks"`
	Signups         int     `gorm:"column:signups;default:0" json:"signups"`
	ActiveReferrals int     `gorm:"column:active_referrals;default:0" json:"activeReferrals"`
	TradeVolume     float64 `gorm:"column:trade_volume;type:decimal(20,8);default:0" json:"tradeVolume"`
	Commissions     float64 `gorm:"column:commissions;type:decimal(20,8);default:0" json:"commissions"`
	Payouts         float64 `gorm:"column:payouts;type:decimal(20,8);default:0" json:"payouts"`
}
