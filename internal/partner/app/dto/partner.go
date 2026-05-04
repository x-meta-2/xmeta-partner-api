package dto

import "xmeta-partner/database"

type PartnerDetail struct {
	Partner            database.Partner      `json:"partner"`
	RecentReferrals    []database.Referral   `json:"recentReferrals"`
	RecentCommissions  []database.Commission `json:"recentCommissions"`
	ReferralLinks      []database.ReferralLink `json:"referralLinks"`
	TotalVolume        float64               `json:"totalVolume"`
	PendingCommissions float64               `json:"pendingCommissions"`
}

type PartnerListResult struct {
	Items []database.Partner `json:"items"`
	Total int                `json:"total"`
}
