package dto

import (
	"time"

	"xmeta-partner/database"
)

type ReferralUserRef struct {
	ID          string `json:"id"`
	MaskedEmail string `json:"maskedEmail"`
	FirstName   string `json:"firstName"`
	LastInitial string `json:"lastInitial"`
	KycLevel    int    `json:"kycLevel"`
}

type ReferralListItem struct {
	ID             string                  `json:"id"`
	PartnerID      string                  `json:"partnerId"`
	ReferredUserID string                  `json:"referredUserId"`
	ReferredUser   *ReferralUserRef        `json:"referredUser"`
	ReferralLinkID *string                 `json:"referralLinkId"`
	Status         database.ReferralStatus `json:"status"`
	StartedAt      time.Time               `json:"startedAt"`
	EndedAt        *time.Time              `json:"endedAt"`
	RegisteredAt   time.Time               `json:"registeredAt"`
	FirstTradeAt *time.Time `json:"firstTradeAt"`
	CreatedAt      time.Time               `json:"createdAt"`
}

type ReferralStats struct {
	Total      int64 `json:"total"`
	Registered int64 `json:"registered"`
	Active     int64 `json:"active"`
	Inactive   int64 `json:"inactive"`
}

type ReferralLinkLookup struct {
	Code            string `json:"code"`
	IsActive        bool   `json:"isActive"`
	PartnerID       string `json:"partnerId"`
	PartnerEmail    string `json:"partnerEmail,omitempty"`
	PartnerFullName string `json:"partnerFullName,omitempty"`
}

type AdminReferralDetail struct {
	Referral    database.Referral     `json:"referral"`
	Commissions []database.Commission `json:"commissions"`
	TotalEarned float64               `json:"totalEarned"`
	TotalVolume float64               `json:"totalVolume"`
}
