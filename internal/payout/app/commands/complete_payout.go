package commands

import (
	"fmt"
	"strings"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/payout/domain"
	"xmeta-partner/internal/payout/port"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type CompletePayoutHandler struct {
	DB      *gorm.DB
	Payouts port.PayoutRepo
}

func (h *CompletePayoutHandler) Handle(id string, params structs.PayoutCompleteParams) (database.Payout, error) {
	var payout database.Payout

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND status = ?", id, database.PayoutStatusProcessing).First(&payout).Error; err != nil {
			return domain.ErrPayoutNotFound
		}

		now := time.Now()
		transactionID := strings.TrimSpace(params.TransactionID)
		if transactionID == "" {
			return fmt.Errorf("transactionId is required")
		}

		payout.Status = database.PayoutStatusCompleted
		payout.TransactionID = transactionID
		payout.ProcessedAt = &now

		if err := tx.Save(&payout).Error; err != nil {
			return err
		}

		return tx.Model(&database.Commission{}).
			Where("payout_id = ?", id).
			Update("status", database.CommissionStatusPaid).Error
	})

	if err != nil {
		return database.Payout{}, err
	}

	if err := h.Payouts.Reload(&payout); err != nil {
		return database.Payout{}, err
	}

	return payout, nil
}
