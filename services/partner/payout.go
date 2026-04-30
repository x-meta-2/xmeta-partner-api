package partner

import (
	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/services"
	"xmeta-partner/structs"
)

type PayoutService struct {
	services.BaseService
}

func (s *PayoutService) List(partnerID string, params structs.PayoutListParams) (structs.PaginationResponse, error) {
	pInput := services.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := s.DB.Model(&database.Payout{}).Where("partner_id = ?", partnerID)
	orm = common.Equal(orm, "status", params.Status)

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var payouts []database.Payout
	if err := orm.
		Order("created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&payouts).Error; err != nil {
		return structs.PaginationResponse{}, err
	}

	return structs.PaginationResponse{Total: total, Items: payouts}, nil
}

func (s *PayoutService) Detail(partnerID string, id string) (map[string]interface{}, error) {
	var payout database.Payout
	if err := s.DB.Where("id = ? AND partner_id = ?", id, partnerID).First(&payout).Error; err != nil {
		return nil, err
	}

	var items []database.PayoutItem
	s.DB.Where("payout_id = ?", id).Find(&items)

	return map[string]interface{}{
		"payout": payout,
		"items":  items,
	}, nil
}

// Pending returns SUM of pending commissions for a partner
func (s *PayoutService) Pending(partnerID string) (map[string]interface{}, error) {
	var pendingAmount float64
	s.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND status = ?", partnerID, "pending").
		Select("COALESCE(SUM(commission_amount), 0)").
		Scan(&pendingAmount)

	var pendingCount int64
	s.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND status = ?", partnerID, "pending").
		Count(&pendingCount)

	return map[string]interface{}{
		"pendingAmount": pendingAmount,
		"pendingCount":  pendingCount,
	}, nil
}
