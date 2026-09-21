package commands

import (
	"fmt"
	"log"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/payout/app/queries"

	"gorm.io/gorm"
)

type ProcessDailyPayoutsHandler struct {
	DB *gorm.DB
}

type partnerPending struct {
	PartnerID string `json:"partnerId"`
}

func (h *ProcessDailyPayoutsHandler) Handle() error {
	today := time.Now().Format("2006-01-02")

	var groups []partnerPending
	if err := h.DB.Model(&database.Commission{}).
		Select("DISTINCT partner_id").
		Where("status = ? AND payout_id IS NULL AND DATE(trade_date) < ?", "pending", today).
		Scan(&groups).Error; err != nil {
		return err
	}

	log.Printf("[PayoutWorker] found %d partner groups with pending commissions", len(groups))

	for _, g := range groups {
		payout, err := h.processPartner(g.PartnerID, today)
		if err != nil {
			log.Printf("[PayoutWorker] error processing partner %s: %v", g.PartnerID, err)
			continue
		}
		if payout != nil {
			log.Printf("[PayoutWorker] created payout for partner %s: amount=%.8f commissions=%d", g.PartnerID, payout.Amount, payout.CommissionCount)
		}
	}

	return nil
}

func (h *ProcessDailyPayoutsHandler) processPartner(partnerID, today string) (*database.Payout, error) {
	var payout database.Payout

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		lockKey := partnerAdvisoryKey(partnerID)
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", lockKey).Error; err != nil {
			return err
		}

		var existingActive int64
		if err := tx.Model(&database.Payout{}).
			Where("partner_id = ? AND status IN ?", partnerID,
				[]string{string(database.PayoutStatusPending), string(database.PayoutStatusProcessing)}).
			Count(&existingActive).Error; err != nil {
			return err
		}
		if existingActive > 0 {
			log.Printf("[PayoutWorker] skipping partner %s: active payout already exists", partnerID)
			return nil
		}

		var lockedIDs []string
		if err := tx.Raw(`
			SELECT id FROM commissions
			WHERE partner_id = ? AND status = 'pending' AND payout_id IS NULL AND DATE(trade_date) < ?
			ORDER BY id
			FOR UPDATE
		`, partnerID, today).Scan(&lockedIDs).Error; err != nil {
			return err
		}
		if len(lockedIDs) == 0 {
			log.Printf("[PayoutWorker] skipping partner %s: no eligible commissions after locking", partnerID)
			return nil
		}

		var pending struct {
			Amount      float64
			Count       int64
			PeriodStart time.Time
			PeriodEnd   time.Time
		}
		if err := tx.Model(&database.Commission{}).
			Where("id IN ?", lockedIDs).
			Select("COALESCE(SUM(rebate_amount), 0) as amount, COUNT(*) as count, MIN(trade_date) as period_start, MAX(trade_date) as period_end").
			Scan(&pending).Error; err != nil {
			return err
		}

		if pending.Amount < queries.MinPayoutAmount {
			log.Printf("[PayoutWorker] skipping partner %s: amount %.8f below minimum threshold", partnerID, pending.Amount)
			return nil
		}

		now := time.Now()
		if pending.PeriodStart.IsZero() {
			pending.PeriodStart = now
		}
		if pending.PeriodEnd.IsZero() {
			pending.PeriodEnd = now
		}

		payout = database.Payout{
			PartnerID:       partnerID,
			Amount:          pending.Amount,
			Currency:        "USDT",
			CommissionCount: int(pending.Count),
			PeriodStart:     pending.PeriodStart,
			PeriodEnd:       pending.PeriodEnd,
			Status:          database.PayoutStatusPending,
		}

		if err := tx.Create(&payout).Error; err != nil {
			return err
		}

		result := tx.Model(&database.Commission{}).
			Where("id IN ?", lockedIDs).
			Updates(map[string]interface{}{
				"payout_id": payout.ID,
				"status":    database.CommissionStatusApproved,
			})
		if result.Error != nil {
			return result.Error
		}
		if int(result.RowsAffected) != len(lockedIDs) {
			return fmt.Errorf("commission count mismatch: expected %d, updated %d", len(lockedIDs), result.RowsAffected)
		}

		var commissions []database.Commission
		if err := tx.Where("payout_id = ?", payout.ID).Find(&commissions).Error; err != nil {
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
			return tx.Create(&items).Error
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	if payout.ID == "" {
		return nil, nil
	}
	return &payout, nil
}
