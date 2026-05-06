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
	PartnerID        string           `gorm:"column:partner_id;not null;index" json:"partnerId"`           // commission авах partner
	Partner          *Partner         `gorm:"foreignKey:PartnerID" json:"partner,omitempty"`
	ReferredUserID   string           `gorm:"column:referred_user_id;not null;index" json:"referredUserId"` // trade хийсэн хэрэглэгч
	ReferredUser     *User            `gorm:"foreignKey:ReferredUserID" json:"referredUser,omitempty"`
	PositionID       string           `gorm:"column:position_id;not null;uniqueIndex" json:"positionId"`    // monorepo position ID (dedup key)
	MarketID         string           `gorm:"column:market_id;not null" json:"marketId"`                    // perp.eth_usdt
	Asset            string           `gorm:"column:asset;not null" json:"asset"`                           // шимтгэлийн валют (USDT, USDC)
	CommissionAmount float64          `gorm:"column:commission_amount;type:decimal(20,8);not null" json:"commissionAmount"` // monorepo-оос ирж байгаа шимтгэл
	VolumeUSD        float64          `gorm:"column:volume_usd;type:decimal(20,8);not null;default:0" json:"volumeUsd"` // арилжааны хэмжээ (tier ахихад хэрэглэнэ)
	CommissionRate   float64          `gorm:"column:commission_rate;type:decimal(5,4);not null" json:"commissionRate"`   // tier-ийн хувь (0.30)
	RebateAmount     float64          `gorm:"column:rebate_amount;type:decimal(20,8);not null" json:"rebateAmount"`       // partner-ийн авах дүн (commission × rate)
	TierID           *string          `gorm:"column:tier_id" json:"tierId"`                                // тооцоолсон үеийн tier
	Status           CommissionStatus `gorm:"column:status;not null;default:pending;index" json:"status"`   // pending → approved → paid
	PayoutID         *string          `gorm:"column:payout_id;index" json:"payoutId"`                       // payout batch-тай холбоно
	TradeDate        time.Time        `gorm:"column:trade_date;not null" json:"tradeDate"`                  // арилжаа хийсэн огноо
}
