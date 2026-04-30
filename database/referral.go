package database

import "time"

// ReferralStatus — lifecycle of a referred user relationship.
//
// Switching is allowed: a user may sit under partner A, then unlink
// (status=unlinked, ended_at set) and re-link to partner B. Past
// commissions stay attributed to whichever partner was active at the
// time of the trade — see commission_engine for the time-bounded lookup.
type ReferralStatus string

const (
	ReferralStatusRegistered ReferralStatus = "registered"
	ReferralStatusDeposited  ReferralStatus = "deposited"
	ReferralStatusActive     ReferralStatus = "active"
	ReferralStatusInactive   ReferralStatus = "inactive"
	ReferralStatusUnlinked   ReferralStatus = "unlinked"
)

type (
	ReferralLink struct {
		Base
		PartnerID     string   `gorm:"column:partner_id;not null;index" json:"partnerId"`
		Partner       *Partner `gorm:"foreignKey:PartnerID" json:"partner"`
		Code          string   `gorm:"column:code;not null;uniqueIndex" json:"code"`
		URL           string   `gorm:"column:url;not null" json:"url"`
		Clicks        int      `gorm:"column:clicks;default:0" json:"clicks"`
		Registrations int      `gorm:"column:registrations;default:0" json:"registrations"`
		IsActive      bool     `gorm:"column:is_active;default:true" json:"isActive"`
	}

	// Referral — partner ↔ user relationship over time. Multiple rows per
	// user are allowed (history of switches). The "one currently active
	// referral per user" invariant is enforced by a partial unique index
	// on `(referred_user_id) WHERE ended_at IS NULL`, created in
	// migrations.go.
	Referral struct {
		Base
		PartnerID      string         `gorm:"column:partner_id;not null;index" json:"partnerId"`
		Partner        *Partner       `gorm:"foreignKey:PartnerID" json:"partner"`
		ReferredUserID string         `gorm:"column:referred_user_id;not null;index" json:"referredUserId"`
		ReferredUser   *User          `gorm:"foreignKey:ReferredUserID" json:"referredUser"`
		ReferralLinkID *string        `gorm:"column:referral_link_id;index" json:"referralLinkId"`
		Status         ReferralStatus `gorm:"column:status;not null;default:registered" json:"status"`
		StartedAt      time.Time      `gorm:"column:started_at;not null" json:"startedAt"`
		EndedAt        *time.Time     `gorm:"column:ended_at;index" json:"endedAt"`
		RegisteredAt   time.Time      `gorm:"column:registered_at;not null" json:"registeredAt"`
		FirstDepositAt *time.Time     `gorm:"column:first_deposit_at" json:"firstDepositAt"`
		FirstTradeAt   *time.Time     `gorm:"column:first_trade_at" json:"firstTradeAt"`
	}
)
