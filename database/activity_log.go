package database

type PartnerActivityLog struct {
	Base
	PartnerID  string `gorm:"column:partner_id;index" json:"partnerId"`
	AdminID    string `gorm:"column:admin_id;index" json:"adminId"`
	Action     string `gorm:"column:action;not null" json:"action"`
	Method     string `gorm:"column:method;not null" json:"method"`
	Path       string `gorm:"column:path;not null" json:"path"`
	StatusCode int    `gorm:"column:status_code" json:"statusCode"`
	IP         string `gorm:"column:ip" json:"ip"`
	UserAgent  string `gorm:"column:user_agent" json:"userAgent"`
	Duration   int64  `gorm:"column:duration" json:"duration"`
}
