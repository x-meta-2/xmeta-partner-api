package structs

import "time"

type CommissionListParams struct {
	PaginationInput
	Status         *string `json:"status"`
	ReferredUserID *string `json:"referredUserId"`
}

type AdminCommissionListParams struct {
	PaginationInput
	Status    *string `json:"status"`
	PartnerID *string `json:"partnerId"`
	Asset     *string `json:"asset"`
	Query     string  `json:"query"`
}

type CommissionBreakdownParams struct {
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}

type TradeEventParams struct {
	UserID           string `json:"userId" binding:"required"`
	AccountID        string `json:"accountId"`
	PositionID       string `json:"positionId" binding:"required"`
	MarketID         string `json:"marketId"`
	CommissionAsset  string `json:"commissionAsset"`
	CommissionAmount string `json:"commissionAmount"`
	VolumeInUSD      string `json:"volumeInUSD"`
	CreatedAt        string `json:"createdAt"`
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

