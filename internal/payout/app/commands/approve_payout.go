package commands

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/payout/domain"
	"xmeta-partner/internal/payout/port"

	"gorm.io/gorm"
)

type ApprovePayoutHandler struct {
	DB      *gorm.DB
	Payouts port.PayoutRepo
}

func (h *ApprovePayoutHandler) Handle(id, adminID string) (database.Payout, error) {
	var payout database.Payout

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND status = ?", id, "pending").First(&payout).Error; err != nil {
			return domain.ErrPayoutNotFound
		}

		payout.Status = database.PayoutStatusProcessing
		payout.ApprovedBy = &adminID
		payout.ProcessedAt = nil

		if err := tx.Save(&payout).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return database.Payout{}, err
	}

	if err := h.Payouts.Reload(&payout); err != nil {
		return database.Payout{}, err
	}

	return payout, nil
}
