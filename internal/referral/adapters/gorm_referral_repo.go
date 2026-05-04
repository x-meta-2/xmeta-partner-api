package adapters

import (
	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type GormReferralRepo struct {
	DB *gorm.DB
}

func (r *GormReferralRepo) FindByID(id string) (*database.Referral, error) {
	var referral database.Referral
	if err := r.DB.
		Preload("ReferredUser").
		Preload("ReferralLink").
		Preload("Partner.User").
		Preload("Partner.Tier").
		Where("id = ?", id).
		First(&referral).Error; err != nil {
		return nil, err
	}
	return &referral, nil
}

func (r *GormReferralRepo) FindByIDAndPartner(id, partnerID string) (*database.Referral, error) {
	var referral database.Referral
	if err := r.DB.
		Preload("ReferredUser").
		Where("id = ? AND partner_id = ?", id, partnerID).
		First(&referral).Error; err != nil {
		return nil, err
	}
	return &referral, nil
}

func (r *GormReferralRepo) FindActiveByUserID(userID string) (*database.Referral, error) {
	var referral database.Referral
	if err := r.DB.
		Where("referred_user_id = ? AND ended_at IS NULL", userID).
		First(&referral).Error; err != nil {
		return nil, err
	}
	return &referral, nil
}

func (r *GormReferralRepo) List(partnerID string, params structs.ReferralListParams) ([]database.Referral, int, error) {
	pInput := common.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := r.DB.Model(&database.Referral{}).
		Where("partner_id = ? AND ended_at IS NULL", partnerID)
	orm = common.Equal(orm, "status", params.Status)

	if params.Query != "" {
		q := "%" + params.Query + "%"
		orm = orm.
			Joins("LEFT JOIN users ON users.id = referrals.referred_user_id").
			Where("users.email ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ?", q, q, q)
	}

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var referrals []database.Referral
	if err := orm.
		Preload("ReferredUser").
		Order("created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&referrals).Error; err != nil {
		return nil, 0, err
	}

	return referrals, total, nil
}

func (r *GormReferralRepo) AdminList(params structs.AdminReferralListParams) ([]database.Referral, int, error) {
	pInput := common.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := r.DB.Model(&database.Referral{}).
		Preload("ReferredUser").
		Preload("ReferralLink").
		Preload("Partner.User").
		Preload("Partner.Tier")

	orm = common.Equal(orm, "partner_id", params.PartnerID)
	orm = common.Equal(orm, "status", params.Status)

	if !params.IncludeHistorical {
		orm = orm.Where("ended_at IS NULL")
	}

	if params.Query != "" {
		q := "%" + params.Query + "%"
		orm = orm.
			Joins("LEFT JOIN users ON users.id = referrals.referred_user_id").
			Where("users.email ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ?", q, q, q)
	}

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var referrals []database.Referral
	if err := orm.
		Order("referrals.created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&referrals).Error; err != nil {
		return nil, 0, err
	}

	return referrals, total, nil
}

func (r *GormReferralRepo) Save(referral *database.Referral) error {
	return r.DB.Save(referral).Error
}

func (r *GormReferralRepo) CountHistory(userID string) int64 {
	var count int64
	r.DB.Model(&database.Referral{}).Where("referred_user_id = ?", userID).Count(&count)
	return count
}
