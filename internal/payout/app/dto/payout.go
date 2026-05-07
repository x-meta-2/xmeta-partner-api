package dto

import "xmeta-partner/database"

type PayoutDetail struct {
	Payout database.Payout      `json:"payout"`
	Items  []database.PayoutItem `json:"items"`
}

type PendingInfo struct {
	PendingBalance  float64 `json:"pendingBalance"`
	PendingCount    int64   `json:"pendingCount"`
	TotalPaid       float64 `json:"totalPaid"`
	LastPayoutDate  *string `json:"lastPayoutDate"`
	MinPayoutAmount float64 `json:"minPayoutAmount"`
}
