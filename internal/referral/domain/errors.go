package domain

import "errors"

var (
	ErrLinkNotFound         = errors.New("referral link not found")
	ErrPartnerNotActive     = errors.New("referral code is not active")
	ErrSelfReferral         = errors.New("cannot link to your own referral code")
	ErrNoActiveReferral     = errors.New("no active referral to unlink")
	ErrMaxLinksReached      = errors.New("maximum referral links per partner reached")
	ErrCodeTaken            = errors.New("referral code is already in use")
	ErrCodeGenerationFailed = errors.New("could not generate a unique referral code; try again")
	ErrReferralNotFound     = errors.New("referral not found")
)
