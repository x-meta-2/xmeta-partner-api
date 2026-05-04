package port

import (
	"xmeta-partner/database"
	"xmeta-partner/structs"
)

type ReferralRepo interface {
	FindByID(id string) (*database.Referral, error)
	FindByIDAndPartner(id, partnerID string) (*database.Referral, error)
	FindActiveByUserID(userID string) (*database.Referral, error)
	List(partnerID string, params structs.ReferralListParams) ([]database.Referral, int, error)
	AdminList(params structs.AdminReferralListParams) ([]database.Referral, int, error)
	Save(referral *database.Referral) error
	CountHistory(userID string) int64
}

type ReferralLinkRepo interface {
	FindByCode(code string) (*database.ReferralLink, error)
	FindByCodeActive(code string) (*database.ReferralLink, error)
	List(partnerID string, params structs.ReferralListParams) ([]database.ReferralLink, int, error)
	Create(link *database.ReferralLink) error
	CodeTaken(code string) (bool, error)
	CountByPartner(partnerID string) (int64, error)
}
