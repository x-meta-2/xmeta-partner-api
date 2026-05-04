package commands

import (
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/payout/domain"
	"xmeta-partner/internal/payout/port"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type RejectPayoutHandler struct {
	DB      *gorm.DB
	Payouts port.PayoutRepo
}

func (h *RejectPayoutHandler) Handle(id, adminID string, params structs.PayoutReviewParams) (database.Payout, error) {
	var payout database.Payout

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND status = ?", id, "pending").First(&payout).Error; err != nil {
			return domain.ErrPayoutNotFound
		}

		now := time.Now()
		payout.Status = database.PayoutStatusFailed
		payout.ApprovedBy = &adminID
		payout.ProcessedAt = &now
		payout.FailureReason = params.FailureReason

		if err := tx.Save(&payout).Error; err != nil {
			return err
		}

		return tx.Model(&database.Commission{}).
			Where("payout_id = ?", id).
			Updates(map[string]interface{}{
				"status":    "pending",
				"payout_id": nil,
			}).Error
	})

	if err != nil {
		return database.Payout{}, err
	}

	if err := h.Payouts.Reload(&payout); err != nil {
		return database.Payout{}, err
	}

	return payout, nil
}
