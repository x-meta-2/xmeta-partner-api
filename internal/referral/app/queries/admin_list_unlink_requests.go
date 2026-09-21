package queries

import (
	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type AdminListUnlinkRequestsHandler struct {
	DB *gorm.DB
}

func (h *AdminListUnlinkRequestsHandler) Handle(params structs.ReferralUnlinkRequestListParams) (structs.PaginationResponse, error) {
	pInput := common.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := h.DB.Model(&database.ReferralUnlinkRequest{}).
		Preload("ReferredUser").
		Preload("Partner.User").
		Preload("Reviewer")

	orm = common.Equal(orm, "partner_id", params.PartnerID)
	orm = common.Equal(orm, "status", params.Status)

	if params.Query != "" {
		q := "%" + params.Query + "%"
		orm = orm.
			Joins("LEFT JOIN users referred_users ON referred_users.id = referral_unlink_requests.referred_user_id").
			Joins("LEFT JOIN partners ON partners.id = referral_unlink_requests.partner_id").
			Joins("LEFT JOIN users partner_users ON partner_users.id = partners.user_id").
			Where(`
				referred_users.email ILIKE ?
				OR referred_users.first_name ILIKE ?
				OR referred_users.last_name ILIKE ?
				OR partner_users.email ILIKE ?
				OR partner_users.first_name ILIKE ?
				OR partner_users.last_name ILIKE ?
				OR referral_unlink_requests.referral_code ILIKE ?
			`, q, q, q, q, q, q, q)
	}

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var requests []database.ReferralUnlinkRequest
	if err := orm.
		Order("referral_unlink_requests.created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&requests).Error; err != nil {
		return structs.PaginationResponse{}, err
	}

	return structs.PaginationResponse{Total: total, Items: requests}, nil
}
