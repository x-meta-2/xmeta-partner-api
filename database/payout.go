package database

import "time"

// PayoutStatus — lifecycle of a payout batch.
type PayoutStatus string

const (
	PayoutStatusPending    PayoutStatus = "pending"
	PayoutStatusProcessing PayoutStatus = "processing"
	PayoutStatusCompleted  PayoutStatus = "completed"
	PayoutStatusFailed     PayoutStatus = "failed"
)

type (
	Payout struct {
		Base
		PartnerID       string       `gorm:"column:partner_id;not null;index" json:"partnerId"`
		Partner         *Partner     `gorm:"foreignKey:PartnerID" json:"partner"`
		Amount          float64      `gorm:"column:amount;type:decimal(20,8);not null" json:"amount"`
		Currency        string       `gorm:"column:currency;not null;default:USDT" json:"currency"`
		CommissionCount int          `gorm:"column:commission_count;not null" json:"commissionCount"`
		PeriodStart     time.Time    `gorm:"column:period_start;not null" json:"periodStart"`
		PeriodEnd       time.Time    `gorm:"column:period_end;not null" json:"periodEnd"`
		Status          PayoutStatus `gorm:"column:status;not null;default:pending" json:"status"`
		ProcessedAt     *time.Time   `gorm:"column:processed_at" json:"processedAt"`
		TransactionID   string       `gorm:"column:transaction_id" json:"transactionId"`
		FailureReason   string       `gorm:"column:failure_reason;type:text" json:"failureReason"`
		ApprovedBy      *string      `gorm:"column:approved_by" json:"approvedBy"`
	}

	PayoutItem struct {
		Base
		PayoutID     string  `gorm:"column:payout_id;not null;index" json:"payoutId"`
		CommissionID string  `gorm:"column:commission_id;not null;index" json:"commissionId"`
		Amount       float64 `gorm:"column:amount;type:decimal(20,8);not null" json:"amount"`
	}
)
