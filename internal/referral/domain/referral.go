package domain

type ReferralStatus string

const (
	StatusRegistered ReferralStatus = "registered"
	StatusActive     ReferralStatus = "active"
	StatusInactive   ReferralStatus = "inactive"
	StatusUnlinked   ReferralStatus = "unlinked"

	MaxReferralLinksPerPartner = 3
)
