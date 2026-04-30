package admin

import (
	"fmt"
	"time"

	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/services"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type PayoutService struct {
	services.BaseService
}

// List returns paginated payouts with status/partner filter, preload Partner
func (s *PayoutService) List(params structs.PayoutListParams) (structs.PaginationResponse, error) {
	pInput := services.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := s.DB.Model(&database.Payout{}).Preload("Partner")
	orm = common.Equal(orm, "status", params.Status)
	orm = common.Equal(orm, "partner_id", params.PartnerID)

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

func (s *PayoutService) Detail(id string) (map[string]interface{}, error) {
	var payout database.Payout
	if err := s.DB.Preload("Partner").Where("id = ?", id).First(&payout).Error; err != nil {
		return nil, err
	}

	var items []database.PayoutItem
	s.DB.Where("payout_id = ?", id).Find(&items)

	return map[string]interface{}{
		"payout": payout,
		"items":  items,
	}, nil
}

func (s *PayoutService) PendingList(params structs.PayoutListParams) (structs.PaginationResponse, error) {
	params.Status = strPtr("pending")
	return s.List(params)
}

// Approve sets payout status=processing, sets approved_by, updates commissions to paid.
// Uses a GORM transaction.
func (s *PayoutService) Approve(id string, adminID string) (database.Payout, error) {
	var payout database.Payout

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND status = ?", id, "pending").First(&payout).Error; err != nil {
			return fmt.Errorf("payout not found or already processed")
		}

		now := time.Now()
		payout.Status = "processing"
		payout.ApprovedBy = &adminID
		payout.ProcessedAt = &now

		if err := tx.Save(&payout).Error; err != nil {
			return err
		}

		if err := tx.Model(&database.Commission{}).Where("payout_id = ?", id).Update("status", "paid").Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return database.Payout{}, err
	}

	s.DB.Preload("Partner").Where("id = ?", id).First(&payout)
	return payout, nil
}

// Reject sets payout status=failed with reason, reverts commissions back to pending.
// Uses a GORM transaction.
func (s *PayoutService) Reject(id string, adminID string, params structs.PayoutReviewParams) (database.Payout, error) {
	var payout database.Payout

	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND status = ?", id, "pending").First(&payout).Error; err != nil {
			return fmt.Errorf("payout not found or already processed")
		}

		now := time.Now()
		payout.Status = "failed"
		payout.ApprovedBy = &adminID
		payout.ProcessedAt = &now
		payout.FailureReason = params.FailureReason

		if err := tx.Save(&payout).Error; err != nil {
			return err
		}

		if err := tx.Model(&database.Commission{}).Where("payout_id = ?", id).Updates(map[string]interface{}{
			"status":    "pending",
			"payout_id": nil,
		}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return database.Payout{}, err
	}

	s.DB.Preload("Partner").Where("id = ?", id).First(&payout)
	return payout, nil
}

func strPtr(s string) *string {
	return &s
}
