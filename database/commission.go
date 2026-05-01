package database

import "time"

// CommissionStatus — lifecycle of a single commission row.
type CommissionStatus string

const (
	CommissionStatusPending   CommissionStatus = "pending"
	CommissionStatusApproved  CommissionStatus = "approved"
	CommissionStatusPaid      CommissionStatus = "paid"
	CommissionStatusCancelled CommissionStatus = "cancelled"
)

type Commission struct {
	Base
	PartnerID        string           `gorm:"column:partner_id;not null;index" json:"partnerId"`
	Partner          *Partner         `gorm:"foreignKey:PartnerID" json:"partner"`
	ReferredUserID   string           `gorm:"column:referred_user_id;not null;index" json:"referredUserId"`
	TradeID          string           `gorm:"column:trade_id;not null;index" json:"tradeId"`
	TradeAmount      float64          `gorm:"column:trade_amount;type:decimal(20,8);not null" json:"tradeAmount"`
	CommissionRate   float64          `gorm:"column:commission_rate;type:decimal(5,4);not null" json:"commissionRate"`
	CommissionAmount float64          `gorm:"column:commission_amount;type:decimal(20,8);not null" json:"commissionAmount"`
	TierID           *string          `gorm:"column:tier_id" json:"tierId"`
	Status           CommissionStatus `gorm:"column:status;not null;default:pending;index" json:"status"`
	PayoutID         *string          `gorm:"column:payout_id;index" json:"payoutId"`
	TradeDate        time.Time        `gorm:"column:trade_date;not null" json:"tradeDate"`
}
