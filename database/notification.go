package database

type PartnerNotification struct {
	Base
	PartnerID string                 `gorm:"column:partner_id;not null;index" json:"partnerId"`
	Title     string                 `gorm:"column:title;not null" json:"title"`
	Message   string                 `gorm:"column:message;type:text;not null" json:"message"`
	Type      string                 `gorm:"column:type;not null" json:"type"`
	IsRead    bool                   `gorm:"column:is_read;default:false" json:"isRead"`
	MetaData  map[string]interface{} `gorm:"serializer:json" json:"metaData"`
}
