package domain

type CommissionStatus string

const (
	StatusPending   CommissionStatus = "pending"
	StatusApproved  CommissionStatus = "approved"
	StatusPaid      CommissionStatus = "paid"
	StatusCancelled CommissionStatus = "cancelled"
)
