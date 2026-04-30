package services

import (
	"fmt"
	"log"
	"time"

	"xmeta-partner/database"

	"gorm.io/gorm"
)

type PayoutWorkerService struct {
	BaseService
}

// ProcessDailyPayouts aggregates pending commissions per partner and creates payout records.
// Full implementation:
//  1. Group pending commissions by partner_id where trade_date < today
//  2. For each partner: create Payout record, create PayoutItems, update commissions to approved with payout_id
//  3. All in a transaction per partner
//  4. Return summary of processed payouts
func (s *PayoutWorkerService) ProcessDailyPayouts() error {
	log.Println("[PayoutWorker] Starting daily payout processing...")

	today := time.Now().Format("2006-01-02")

	// Get all partners with pending commissions where trade_date < today
	type PartnerPending struct {
		PartnerID       string  `json:"partnerId"`
		TotalAmount     float64 `json:"totalAmount"`
		CommissionCount int     `json:"commissionCount"`
	}

	var pending []PartnerPending
	if err := s.DB.Model(&database.Commission{}).
		Where("status = ? AND payout_id IS NULL AND DATE(trade_date) < ?", "pending", today).
		Select("partner_id, SUM(commission_amount) as total_amount, COUNT(*) as commission_count").
		Group("partner_id").
		Having("SUM(commission_amount) > 0").
		Find(&pending).Error; err != nil {
		return fmt.Errorf("failed to query pending commissions: %w", err)
	}

	if len(pending) == 0 {
		log.Println("[PayoutWorker] No pending commissions found")
		return nil
	}

	log.Printf("[PayoutWorker] Found %d partners with pending commissions", len(pending))

	now := time.Now()
	periodEnd := now
	periodStart := now.AddDate(0, 0, -1)

	processedCount := 0
	skippedCount := 0

	for _, p := range pending {
		// Minimum payout threshold ($1)
		if p.TotalAmount < 1.0 {
			log.Printf("[PayoutWorker] Skipping partner %s: amount %.8f below minimum", p.PartnerID, p.TotalAmount)
			skippedCount++
			continue
		}

		err := s.DB.Transaction(func(tx *gorm.DB) error {
			// Create payout record
			payout := database.Payout{
				PartnerID:       p.PartnerID,
				Amount:          p.TotalAmount,
				Currency:        "USDT",
				CommissionCount: p.CommissionCount,
				PeriodStart:     periodStart,
				PeriodEnd:       periodEnd,
				Status:          "pending",
			}

			if err := tx.Create(&payout).Error; err != nil {
				return fmt.Errorf("error creating payout: %w", err)
			}

			// Link commissions to this payout and update status to approved
			if err := tx.Model(&database.Commission{}).
				Where("partner_id = ? AND status = ? AND payout_id IS NULL AND DATE(trade_date) < ?", p.PartnerID, "pending", today).
				Updates(map[string]interface{}{
					"payout_id": payout.ID,
					"status":    "approved",
				}).Error; err != nil {
				return fmt.Errorf("error linking commissions: %w", err)
			}

			// Create payout items for audit trail
			var commissions []database.Commission
			if err := tx.Where("payout_id = ?", payout.ID).Find(&commissions).Error; err != nil {
				return fmt.Errorf("error fetching linked commissions: %w", err)
			}

			for _, c := range commissions {
				item := database.PayoutItem{
					PayoutID:     payout.ID,
					CommissionID: c.ID,
					Amount:       c.CommissionAmount,
				}
				if err := tx.Create(&item).Error; err != nil {
					return fmt.Errorf("error creating payout item: %w", err)
				}
			}

			log.Printf("[PayoutWorker] Created payout %s for partner %s: $%.8f (%d commissions)",
				payout.ID, p.PartnerID, p.TotalAmount, len(commissions))

			return nil
		})

		if err != nil {
			log.Printf("[PayoutWorker] Error processing payout for partner %s: %v", p.PartnerID, err)
			continue
		}

		processedCount++
	}

	log.Printf("[PayoutWorker] Daily payout processing completed: %d processed, %d skipped", processedCount, skippedCount)
	return nil
}
