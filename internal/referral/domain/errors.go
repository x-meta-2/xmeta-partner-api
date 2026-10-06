package domain

import "errors"

var (
	ErrLinkNotFound          = errors.New("referral link not found")
	ErrPartnerNotActive      = errors.New("referral code is not active")
	ErrSelfReferral          = errors.New("cannot link to your own referral code")
	ErrCircularReferral      = errors.New("cannot link to a partner in your referral network")
	ErrActiveReferralExists  = errors.New("user is already linked to an active partner")
	ErrNoActiveReferral      = errors.New("no active referral to unlink")
	ErrUnlinkCooldown        = errors.New("referral can only be unlinked once every 7 days")
	ErrMaxLinksReached       = errors.New("maximum referral links per partner reached")
	ErrCodeTaken             = errors.New("referral code is already in use")
	ErrCodeGenerationFailed  = errors.New("could not generate a unique referral code; try again")
	ErrReferralNotFound      = errors.New("referral not found")
	ErrPendingUnlinkRequest  = errors.New("pending unlink request already exists")
	ErrUnlinkRequestNotFound = errors.New("referral unlink request not found")
	ErrUnlinkRequestReviewed = errors.New("referral unlink request is already reviewed")
)
