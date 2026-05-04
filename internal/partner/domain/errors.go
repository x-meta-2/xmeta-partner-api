package domain

import "errors"

var (
	ErrPartnerNotFound       = errors.New("partner not found")
	ErrApplicationNotFound   = errors.New("application not found or already reviewed")
	ErrTierNotFound          = errors.New("tier not found")
	ErrDefaultTierMissing    = errors.New("no default partner tier configured")
	ErrTierInUse             = errors.New("cannot delete tier: partners are assigned to it")
	ErrReferralCodeCollision = errors.New("could not generate a unique referral code; retry")
)
