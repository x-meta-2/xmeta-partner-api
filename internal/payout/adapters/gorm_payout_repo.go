package adapters

import (
	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type GormPayoutRepo struct {
	DB *gorm.DB
}

func (r *GormPayoutRepo) FindPendingByID(id string) (*database.Payout, error) {
	var payout database.Payout
	if err := r.DB.Where("id = ? AND status = ?", id, "pending").First(&payout).Error; err != nil {
		return nil, err
	}
	return &payout, nil
}

func (r *GormPayoutRepo) List(partnerID string, params structs.PayoutListParams) ([]database.Payout, int, error) {
	pInput := common.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := r.DB.Model(&database.Payout{}).Where("partner_id = ?", partnerID)
	orm = common.Equal(orm, "status", params.Status)

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var payouts []database.Payout
	if err := orm.
		Order("created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&payouts).Error; err != nil {
		return nil, 0, err
	}

	return payouts, total, nil
}

func (r *GormPayoutRepo) AdminList(params structs.PayoutListParams) ([]database.Payout, int, error) {
	pInput := common.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := r.DB.Model(&database.Payout{}).Preload("Partner")
	orm = common.Equal(orm, "status", params.Status)
	orm = common.Equal(orm, "partner_id", params.PartnerID)

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var payouts []database.Payout
	if err := orm.
		Order("created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&payouts).Error; err != nil {
		return nil, 0, err
	}

	return payouts, total, nil
}

func (r *GormPayoutRepo) Detail(partnerID, id string) (*database.Payout, []database.PayoutItem, error) {
	var payout database.Payout
	if err := r.DB.Where("id = ? AND partner_id = ?", id, partnerID).First(&payout).Error; err != nil {
		return nil, nil, err
	}

	var items []database.PayoutItem
	if err := r.DB.Where("payout_id = ?", id).Find(&items).Error; err != nil {
		return nil, nil, err
	}

	return &payout, items, nil
}

func (r *GormPayoutRepo) AdminDetail(id string) (*database.Payout, []database.PayoutItem, error) {
	var payout database.Payout
	if err := r.DB.Preload("Partner").Where("id = ?", id).First(&payout).Error; err != nil {
		return nil, nil, err
	}

	var items []database.PayoutItem
	if err := r.DB.Where("payout_id = ?", id).Find(&items).Error; err != nil {
		return nil, nil, err
	}

	return &payout, items, nil
}

func (r *GormPayoutRepo) Save(payout *database.Payout) error {
	return r.DB.Save(payout).Error
}

func (r *GormPayoutRepo) Create(payout *database.Payout) error {
	return r.DB.Create(payout).Error
}

func (r *GormPayoutRepo) Reload(payout *database.Payout) error {
	return r.DB.Preload("Partner").Where("id = ?", payout.ID).First(payout).Error
}
