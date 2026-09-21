package database

import "time"

type ReferralUnlinkRequestStatus string

const (
	ReferralUnlinkRequestStatusPending  ReferralUnlinkRequestStatus = "pending"
	ReferralUnlinkRequestStatusApproved ReferralUnlinkRequestStatus = "approved"
	ReferralUnlinkRequestStatusRejected ReferralUnlinkRequestStatus = "rejected"
)

type ReferralUnlinkRequest struct {
	Base
	ReferralID     string                      `gorm:"column:referral_id;not null;index" json:"referralId"`
	Referral       *Referral                   `gorm:"foreignKey:ReferralID" json:"referral"`
	PartnerID      string                      `gorm:"column:partner_id;not null;index" json:"partnerId"`
	Partner        *Partner                    `gorm:"foreignKey:PartnerID" json:"partner"`
	ReferredUserID string                      `gorm:"column:referred_user_id;not null;index" json:"referredUserId"`
	ReferredUser   *User                       `gorm:"foreignKey:ReferredUserID" json:"referredUser"`
	ReferralCode   string                      `gorm:"column:referral_code;not null" json:"referralCode"`
	Reason         string                      `gorm:"column:reason;type:text;not null" json:"reason"`
	Status         ReferralUnlinkRequestStatus `gorm:"column:status;not null;default:pending;index" json:"status"`
	ReviewedBy     *string                     `gorm:"column:reviewed_by;index" json:"reviewedBy"`
	Reviewer       *AdminUser                  `gorm:"foreignKey:ReviewedBy;references:ID" json:"reviewer"`
	AdminNote      string                      `gorm:"column:admin_note;type:text" json:"adminNote"`
	ReviewedAt     *time.Time                  `gorm:"column:reviewed_at" json:"reviewedAt"`
}
