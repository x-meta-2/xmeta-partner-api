package domain

import "errors"

var (
	ErrPartnerNotFound    = errors.New("partner not found or inactive")
	ErrNoTierAssigned     = errors.New("partner has no tier assigned")
	ErrNoActiveReferral   = errors.New("no active referral for trade attribution")
	ErrUserNotKycVerified = errors.New("referred user has not passed KYC")
)
