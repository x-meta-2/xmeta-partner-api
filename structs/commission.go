package structs

import "time"

type CommissionListParams struct {
	PaginationInput
	Status         *string `json:"status"`
	ReferredUserID *string `json:"referredUserId"`
}

type CommissionBreakdownParams struct {
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}

type TradeEventParams struct {
	TradeID        string  `json:"tradeId" binding:"required"`
	UserID         string  `json:"userId" binding:"required"`
	TradeAmount    float64 `json:"tradeAmount" binding:"required"`
	TradeFee       float64 `json:"tradeFee" binding:"required"`
	Symbol         string  `json:"symbol"`
	TradeTimestamp int64   `json:"tradeTimestamp"`
}

// UserRegisteredParams — body of `POST /internal/link-referral`. Only
// the two fields needed to attach a referral are accepted; identity and
// device metadata stay on the user/audit tables in xmeta-monorepo.
type UserRegisteredParams struct {
	UserID       string `json:"userId" binding:"required"`
	ReferralCode string `json:"referralCode" binding:"required"`
}

// UnlinkReferralEventParams — body of `POST /internal/unlink-referral`.
// Used by xmeta-monorepo for server-side detachments (account closure,
// compliance flags, etc.). Plain users still go through the Bearer-auth
// `/partner/auth/unlink-referral` endpoint.
type UnlinkReferralEventParams struct {
	UserID string `json:"userId" binding:"required"`
}

type UserDepositedParams struct {
	UserID string  `json:"userId" binding:"required"`
	Amount float64 `json:"amount" binding:"required"`
}
