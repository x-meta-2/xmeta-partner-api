package dto

import "xmeta-partner/database"

type PayoutDetail struct {
	Payout database.Payout      `json:"payout"`
	Items  []database.PayoutItem `json:"items"`
}

type PendingInfo struct {
	PendingAmount float64 `json:"pendingAmount"`
	PendingCount  int64   `json:"pendingCount"`
}
