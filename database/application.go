package database

import "time"

// ApplicationStatus — partner application review lifecycle.
type ApplicationStatus string

const (
	ApplicationStatusPending  ApplicationStatus = "pending"
	ApplicationStatusApproved ApplicationStatus = "approved"
	ApplicationStatusRejected ApplicationStatus = "rejected"
)

// PartnerApplication — partner-д элсэхийг хүссэн user-ийн хүсэлт.
// UserID = users.id = Cognito sub. Email/name зэрэг identity field-үүд нь
// User table-аас preload-оор татагдана.
type PartnerApplication struct {
	Base
	UserID          string                 `gorm:"column:user_id;not null;index" json:"userId"`
	User            *User                  `gorm:"foreignKey:UserID" json:"user"`
	CompanyName     string                 `gorm:"column:company_name" json:"companyName"`
	Website         string                 `gorm:"column:website" json:"website"`
	SocialMedia     map[string]interface{} `gorm:"column:social_media;serializer:json" json:"socialMedia"`
	AudienceSize    string                 `gorm:"column:audience_size" json:"audienceSize"`
	PromotionPlan   string                 `gorm:"column:promotion_plan;type:text" json:"promotionPlan"`
	Status          ApplicationStatus      `gorm:"column:status;not null;default:pending;index" json:"status"`
	ReviewedBy      *string                `gorm:"column:reviewed_by" json:"reviewedBy"`
	ReviewedAt      *time.Time             `gorm:"column:reviewed_at" json:"reviewedAt"`
	RejectionReason string                 `gorm:"column:rejection_reason;type:text" json:"rejectionReason"`
}
