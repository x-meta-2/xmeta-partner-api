package domain

import (
	"errors"
	"time"
)

var (
	ErrLinkNotFound          = errors.New("referral link not found")
	ErrPartnerNotActive      = errors.New("referral code is not active")
	ErrSelfReferral          = errors.New("cannot link to your own referral code")
	ErrCircularReferral      = errors.New("cannot link to a partner in your referral network")
	ErrActiveReferralExists  = errors.New("user is already linked to an active partner")
	ErrLinkCooldown          = errors.New("referral can only be linked 7 days after unlinking")
	ErrNoActiveReferral      = errors.New("no active referral to unlink")
	ErrMaxLinksReached       = errors.New("maximum referral links per partner reached")
	ErrCodeTaken             = errors.New("referral code is already in use")
	ErrCodeGenerationFailed  = errors.New("could not generate a unique referral code; try again")
	ErrReferralNotFound      = errors.New("referral not found")
	ErrPendingUnlinkRequest  = errors.New("pending unlink request already exists")
	ErrUnlinkRequestNotFound = errors.New("referral unlink request not found")
	ErrUnlinkRequestReviewed = errors.New("referral unlink request is already reviewed")
)

type LinkCooldownError struct {
	NextLinkAt time.Time
}

func (e LinkCooldownError) Error() string {
	return ErrLinkCooldown.Error()
}

func (e LinkCooldownError) Is(target error) bool {
	return target == ErrLinkCooldown
}
