package domain

type PartnerStatus string

const (
	StatusPending   PartnerStatus = "pending"
	StatusActive    PartnerStatus = "active"
	StatusSuspended PartnerStatus = "suspended"
)

type ApplicationStatus string

const (
	ApplicationPending  ApplicationStatus = "pending"
	ApplicationApproved ApplicationStatus = "approved"
	ApplicationRejected ApplicationStatus = "rejected"
)
