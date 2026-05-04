package domain

import "errors"

var (
	ErrPayoutNotFound = errors.New("payout not found or already processed")
)
