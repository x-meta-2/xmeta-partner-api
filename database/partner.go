package database

// PartnerStatus — lifecycle of a partner row.
type PartnerStatus string

const (
	PartnerStatusPending   PartnerStatus = "pending"
	PartnerStatusActive    PartnerStatus = "active"
	PartnerStatusSuspended PartnerStatus = "suspended"
)

// Partner — 1:1 extension of User for partner-role data.
//
// partners.id is a separate UUID (auto-generated via Base.BeforeCreate) —
// preserved across suspend/rejoin cycles for audit trail in commissions,
// payouts, referrals, etc.
//
// partners.user_id = users.id = Cognito `sub`. This is the natural key
// used by PartnerAuth middleware to resolve a login token to a partner.
//
// Identity fields (email, first_name, last_name) live on User and are
// preloaded via the `User` relationship — never duplicated here.
type Partner struct {
	Base
	UserID         string                 `gorm:"column:user_id;not null;uniqueIndex" json:"userId"`
	User           *User                  `gorm:"foreignKey:UserID" json:"user"`
	TierID         string                 `gorm:"column:tier_id;index" json:"tierId"`
	Tier           *PartnerTier           `gorm:"foreignKey:TierID" json:"tier"`
	Status         PartnerStatus          `gorm:"column:status;not null;default:pending;index" json:"status"`
	ReferralCode   string                 `gorm:"column:referral_code;not null;uniqueIndex" json:"referralCode"`
	TotalReferrals int                    `gorm:"column:total_referrals;default:0" json:"totalReferrals"`
	TotalEarnings  float64                `gorm:"column:total_earnings;type:decimal(20,8);default:0" json:"totalEarnings"`
	CompanyName    string                 `gorm:"column:company_name" json:"companyName"`
	Website        string                 `gorm:"column:website" json:"website"`
	// SocialMedia mirrors PartnerApplication.SocialMedia at approve time.
	// Application stays as the review snapshot; this column is the
	// partner's *current* profile (Settings edits land here).
	//
	// `serializer:json` round-trips `map[string]interface{}` through
	// encoding/json so the column can stay TEXT — works with both lib/pq
	// and pgx. Partial `Updates(map)` bypass the serializer, so the
	// service marshals the value to a JSON string before assigning.
	SocialMedia map[string]interface{} `gorm:"column:social_media;serializer:json" json:"socialMedia"`
	MetaData    map[string]interface{} `gorm:"column:meta_data;serializer:json" json:"metaData"`
}
