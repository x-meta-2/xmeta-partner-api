package adapters

import (
	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type GormReferralLinkRepo struct {
	DB *gorm.DB
}

func (r *GormReferralLinkRepo) FindByCode(code string) (*database.ReferralLink, error) {
	var link database.ReferralLink
	if err := r.DB.Where("code = ?", code).First(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *GormReferralLinkRepo) FindByCodeActive(code string) (*database.ReferralLink, error) {
	var link database.ReferralLink
	if err := r.DB.Where("code = ? AND is_active = ?", code, true).First(&link).Error; err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *GormReferralLinkRepo) List(partnerID string, params structs.ReferralListParams) ([]database.ReferralLink, int, error) {
	pInput := common.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := r.DB.Model(&database.ReferralLink{}).Where("partner_id = ?", partnerID)

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var links []database.ReferralLink
	if err := orm.
		Order("created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&links).Error; err != nil {
		return nil, 0, err
	}

	for i := range links {
		var live int64
		r.DB.Model(&database.Referral{}).
			Where("referral_link_id = ? AND ended_at IS NULL", links[i].ID).
			Count(&live)
		links[i].Registrations = int(live)
	}

	return links, total, nil
}

func (r *GormReferralLinkRepo) Create(link *database.ReferralLink) error {
	return r.DB.Create(link).Error
}

func (r *GormReferralLinkRepo) CodeTaken(code string) (bool, error) {
	var linkCount int64
	if err := r.DB.Model(&database.ReferralLink{}).Where("code = ?", code).Count(&linkCount).Error; err != nil {
		return false, err
	}
	if linkCount > 0 {
		return true, nil
	}

	var partnerCount int64
	if err := r.DB.Model(&database.Partner{}).Where("referral_code = ?", code).Count(&partnerCount).Error; err != nil {
		return false, err
	}
	return partnerCount > 0, nil
}

func (r *GormReferralLinkRepo) CountByPartner(partnerID string) (int64, error) {
	var count int64
	if err := r.DB.Model(&database.ReferralLink{}).Where("partner_id = ?", partnerID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
