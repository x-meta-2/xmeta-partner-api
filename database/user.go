package database

// User — read-only reference to the shared `users` table.
//
// The `users` table is owned and populated by xmeta-monorepo (via DynamoToSql
// sync from DynamoDB `xmeta-users`). `users.id` equals the Cognito `sub` claim.
//
// Partner-api must ONLY read from this table; never insert/update/delete.
// It is intentionally excluded from AutoMigrate in database/client.go.
type User struct {
	Base
	Email              string                 `gorm:"column:email;not null;uniqueIndex" json:"email"`
	BinanceEmail       string                 `gorm:"column:binance_email" json:"binanceEmail"`
	FirstName          string                 `gorm:"column:first_name" json:"firstName"`
	LastName           string                 `gorm:"column:last_name" json:"lastName"`
	SubAccountID       string                 `gorm:"column:sub_account_id" json:"subAccountId"`
	CanTrade           bool                   `gorm:"column:can_trade;default:false" json:"canTrade"`
	CanWithdraw        bool                   `gorm:"column:can_withdraw;default:false" json:"canWithdraw"`
	IsWhitelistEnabled bool                   `gorm:"column:is_whitelist_enabled;default:false" json:"isWhitelistEnabled"`
	KycLevel           int                    `gorm:"column:kyc_level;default:0" json:"kycLevel"`
	VipLevel           int                    `gorm:"column:vip_level;default:0" json:"vipLevel"`
	Status             int                    `gorm:"column:status;default:1" json:"status"`
	MetaData           map[string]interface{} `gorm:"serializer:json" json:"metaData"`
}

func (User) TableName() string {
	return "users"
}
