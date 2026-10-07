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
		PartnerID             string                 `gorm:"column:partner_id;not null;index" json:"partnerId"`
		Partner               *Partner               `gorm:"foreignKey:PartnerID" json:"partner"`
		Amount                float64                `gorm:"column:amount;type:decimal(20,8);not null" json:"amount"` // payable amount; commission rows are already net of tax
		GrossAmount           float64                `gorm:"column:gross_amount;type:decimal(20,8);not null;default:0" json:"grossAmount"`
		TaxRate               float64                `gorm:"column:tax_rate;type:decimal(5,4);not null;default:0" json:"taxRate"`
		TaxAmount             float64                `gorm:"column:tax_amount;type:decimal(20,8);not null;default:0" json:"taxAmount"`
		Currency              string                 `gorm:"column:currency;not null;default:USDT" json:"currency"`
		CommissionCount       int                    `gorm:"column:commission_count;not null" json:"commissionCount"`
		PeriodStart           time.Time              `gorm:"column:period_start;not null" json:"periodStart"`
		PeriodEnd             time.Time              `gorm:"column:period_end;not null" json:"periodEnd"`
		Status                PayoutStatus           `gorm:"column:status;not null;default:pending" json:"status"`
		ProcessedAt           *time.Time             `gorm:"column:processed_at" json:"processedAt"`
		TransactionID         string                 `gorm:"column:transaction_id" json:"transactionId"`
		TransferService       string                 `gorm:"column:transfer_service" json:"transferService"`
		TransferRequestID     string                 `gorm:"column:transfer_request_id;index" json:"transferRequestId"`
		TransferTransactionID string                 `gorm:"column:transfer_transaction_id;index" json:"transferTransactionId"`
		TransferStatus        string                 `gorm:"column:transfer_status;not null;default:not_started" json:"transferStatus"`
		TransferResponse      map[string]interface{} `gorm:"column:transfer_response;serializer:json" json:"transferResponse"`
		TransferError         string                 `gorm:"column:transfer_error;type:text" json:"transferError"`
		FailureReason         string                 `gorm:"column:failure_reason;type:text" json:"failureReason"`
		ApprovedBy            *string                `gorm:"column:approved_by" json:"approvedBy"`
	}

	PayoutItem struct {
		Base
		PayoutID     string      `gorm:"column:payout_id;not null;index" json:"payoutId"`
		CommissionID string      `gorm:"column:commission_id;not null;index" json:"commissionId"`
		Commission   *Commission `gorm:"foreignKey:CommissionID" json:"commission,omitempty"`
		Amount       float64     `gorm:"column:amount;type:decimal(20,8);not null" json:"amount"`
	}
)
