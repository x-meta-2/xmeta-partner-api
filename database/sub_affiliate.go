package database

import "time"

type SubAffiliateInvite struct {
	Base
	PartnerID    string     `gorm:"column:partner_id;not null;index" json:"partnerId"`
	Partner      *Partner   `gorm:"foreignKey:PartnerID" json:"partner"`
	Email        string     `gorm:"column:email;not null" json:"email"`
	InviteCode   string     `gorm:"column:invite_code;not null;uniqueIndex" json:"inviteCode"`
	Status       string     `gorm:"column:status;not null;default:pending" json:"status"`
	OverrideRate float64    `gorm:"column:override_rate;type:decimal(5,4);default:0.1000" json:"overrideRate"`
	ExpiresAt    *time.Time `gorm:"column:expires_at" json:"expiresAt"`
	AcceptedAt   *time.Time `gorm:"column:accepted_at" json:"acceptedAt"`
	SubPartnerID *string    `gorm:"column:sub_partner_id" json:"subPartnerId"`
}
