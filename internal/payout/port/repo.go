package port

import (
	"xmeta-partner/database"
	"xmeta-partner/structs"
)

type PayoutRepo interface {
	FindPendingByID(id string) (*database.Payout, error)
	List(partnerID string, params structs.PayoutListParams) ([]database.Payout, int, error)
	AdminList(params structs.PayoutListParams) ([]database.Payout, int, error)
	Detail(partnerID, id string) (*database.Payout, []database.PayoutItem, error)
	AdminDetail(id string) (*database.Payout, []database.PayoutItem, error)
	Save(payout *database.Payout) error
	Create(payout *database.Payout) error
	Reload(payout *database.Payout) error
}

type PayoutItemRepo interface {
	CreateBatch(items []database.PayoutItem) error
}
