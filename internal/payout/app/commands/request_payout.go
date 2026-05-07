package commands

import (
	"fmt"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/payout/app/queries"

	"gorm.io/gorm"
)

type RequestPayoutHandler struct {
	DB *gorm.DB
}

func (h *RequestPayoutHandler) Handle(partnerID string) (*database.Payout, error) {
	var payout database.Payout

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		// Advisory lock per partner — serialises concurrent requests from the
		// same partner without blocking other partners.
		lockKey := partnerAdvisoryKey(partnerID)
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", lockKey).Error; err != nil {
			return err
		}

		var existingPending int64
		if err := tx.Model(&database.Payout{}).
			Where("partner_id = ? AND status IN ?", partnerID,
				[]string{string(database.PayoutStatusPending), string(database.PayoutStatusProcessing)}).
			Count(&existingPending).Error; err != nil {
			return err
		}
		if existingPending > 0 {
			return fmt.Errorf("you already have a pending payout request")
		}

		// Lock the exact eligible commission rows so a concurrent tx blocks
		// instead of reading the same set.
		var lockedIDs []string
		if err := tx.Raw(`
			SELECT id FROM commissions
			WHERE partner_id = ? AND status = 'pending' AND payout_id IS NULL
			ORDER BY id
			FOR UPDATE
		`, partnerID).Scan(&lockedIDs).Error; err != nil {
			return err
		}

		if len(lockedIDs) == 0 {
			return fmt.Errorf("no eligible commissions")
		}

		var pending struct {
			Amount float64
			Count  int64
		}
		if err := tx.Model(&database.Commission{}).
			Where("id IN ?", lockedIDs).
			Select("COALESCE(SUM(rebate_amount), 0) as amount, COUNT(*) as count").
			Scan(&pending).Error; err != nil {
			return err
		}

		if pending.Amount < queries.MinPayoutAmount {
			return fmt.Errorf("minimum payout amount is $%.0f", queries.MinPayoutAmount)
		}

		now := time.Now()
		payout = database.Payout{
			PartnerID:       partnerID,
			Amount:          pending.Amount,
			Currency:        "USDT",
			CommissionCount: int(pending.Count),
			PeriodStart:     now,
			PeriodEnd:       now,
			Status:          database.PayoutStatusPending,
		}
		if err := tx.Create(&payout).Error; err != nil {
			return err
		}

		// Update only the locked rows — not a second WHERE scan.
		result := tx.Model(&database.Commission{}).
			Where("id IN ?", lockedIDs).
			Updates(map[string]any{
				"payout_id": payout.ID,
				"status":    "approved",
			})
		if result.Error != nil {
			return result.Error
		}

		// Verify every locked row was claimed.
		if int(result.RowsAffected) != len(lockedIDs) {
			return fmt.Errorf("commission count mismatch: expected %d, updated %d", len(lockedIDs), result.RowsAffected)
		}

		var commissions []database.Commission
		if err := tx.Where("id IN ?", lockedIDs).Find(&commissions).Error; err != nil {
			return err
		}

		items := make([]database.PayoutItem, 0, len(commissions))
		for _, c := range commissions {
			items = append(items, database.PayoutItem{
				PayoutID:     payout.ID,
				CommissionID: c.ID,
				Amount:       c.RebateAmount,
			})
		}
		if len(items) > 0 {
			if err := tx.Create(&items).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &payout, nil
}

// partnerAdvisoryKey produces a stable int64 from the partner UUID for
// pg_advisory_xact_lock. FNV-style hash keeps collisions negligible.
func partnerAdvisoryKey(partnerID string) int64 {
	var h uint64 = 14695981039346656037
	for i := 0; i < len(partnerID); i++ {
		h ^= uint64(partnerID[i])
		h *= 1099511628211
	}
	return int64(h)
}
