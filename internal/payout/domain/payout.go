package domain

type PayoutStatus string

const (
	StatusPending    PayoutStatus = "pending"
	StatusProcessing PayoutStatus = "processing"
	StatusCompleted  PayoutStatus = "completed"
	StatusFailed     PayoutStatus = "failed"
)
