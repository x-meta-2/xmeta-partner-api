package dto

import (
	"time"

	"xmeta-partner/database"
)

type ReferralListItem struct {
	ID             string                  `json:"id"`
	PartnerID      string                  `json:"partnerId"`
	ReferredUserID string                  `json:"referredUserId"`
	ReferralLinkID *string                 `json:"referralLinkId"`
	Status         database.ReferralStatus `json:"status"`
	StartedAt      time.Time               `json:"startedAt"`
	EndedAt        *time.Time              `json:"endedAt"`
	RegisteredAt   time.Time               `json:"registeredAt"`
	FirstTradeAt   *time.Time              `json:"firstTradeAt"`
	CreatedAt      time.Time               `json:"createdAt"`
}

type ReferralStats struct {
	Total      int64 `json:"total"`
	Registered int64 `json:"registered"`
	Active     int64 `json:"active"`
	Inactive   int64 `json:"inactive"`
	Unlinked   int64 `json:"unlinked"`
}

type ReferralLinkLookup struct {
	Code            string `json:"code"`
	IsActive        bool   `json:"isActive"`
	PartnerID       string `json:"partnerId"`
	PartnerEmail    string `json:"partnerEmail,omitempty"`
	PartnerFullName string `json:"partnerFullName,omitempty"`
}

type CurrentReferralPartner struct {
	Email        string `json:"email,omitempty"`
	FirstName    string `json:"firstName,omitempty"`
	ReferralCode string `json:"referralCode"`
}

type CurrentReferral struct {
	ReferralCode string                 `json:"referralCode"`
	Partner      CurrentReferralPartner `json:"partner"`
	CanUnlink    bool                   `json:"canUnlink"`
	NextUnlinkAt *time.Time             `json:"nextUnlinkAt,omitempty"`
	NextLinkAt   *time.Time             `json:"nextLinkAt,omitempty"`
}

type ReferralUnlinkResult struct {
	ReferralID   string                  `json:"referralId"`
	Status       database.ReferralStatus `json:"status"`
	UnlinkedAt   time.Time               `json:"unlinkedAt"`
	NextLinkAt   time.Time               `json:"nextLinkAt"`
	NextUnlinkAt time.Time               `json:"nextUnlinkAt"` // deprecated: use nextLinkAt
}

type LinkCooldownPayload struct {
	Code       string    `json:"code"`
	Message    string    `json:"message"`
	NextLinkAt time.Time `json:"nextLinkAt"`
}

type AdminReferralDetail struct {
	Referral    database.Referral     `json:"referral"`
	Commissions []database.Commission `json:"commissions"`
	TotalEarned float64               `json:"totalEarned"`
	TotalVolume float64               `json:"totalVolume"`
}
