package commands

import (
	"log"
	"time"

	"xmeta-partner/database"

	"gorm.io/gorm"
)

type ProcessDailyPayoutsHandler struct {
	DB *gorm.DB
}

type partnerPending struct {
	PartnerID       string  `json:"partnerId"`
	TotalAmount     float64 `json:"totalAmount"`
	CommissionCount int     `json:"commissionCount"`
}

func (h *ProcessDailyPayoutsHandler) Handle() error {
	today := time.Now().Format("2006-01-02")

	var groups []partnerPending
	if err := h.DB.Model(&database.Commission{}).
		Select("partner_id, SUM(commission_amount) as total_amount, COUNT(*) as commission_count").
		Where("status = ? AND payout_id IS NULL AND DATE(trade_date) < ?", "pending", today).
		Group("partner_id").
		Having("SUM(commission_amount) > 0").
		Scan(&groups).Error; err != nil {
		return err
	}

	log.Printf("[PayoutWorker] found %d partner groups with pending commissions", len(groups))

	for _, g := range groups {
		if g.TotalAmount < 1.0 {
			log.Printf("[PayoutWorker] skipping partner %s: amount %.8f below minimum threshold", g.PartnerID, g.TotalAmount)
			continue
		}

		if err := h.processPartner(g, today); err != nil {
			log.Printf("[PayoutWorker] error processing partner %s: %v", g.PartnerID, err)
			continue
		}

		log.Printf("[PayoutWorker] created payout for partner %s: amount=%.8f commissions=%d", g.PartnerID, g.TotalAmount, g.CommissionCount)
	}

	return nil
}

func (h *ProcessDailyPayoutsHandler) processPartner(g partnerPending, today string) error {
	return h.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		payout := database.Payout{
			PartnerID:       g.PartnerID,
			Amount:          g.TotalAmount,
			Currency:        "USDT",
			CommissionCount: g.CommissionCount,
			PeriodStart:     now.AddDate(0, 0, -1),
			PeriodEnd:       now,
			Status:          database.PayoutStatusPending,
		}

		if err := tx.Create(&payout).Error; err != nil {
			return err
		}

		if err := tx.Model(&database.Commission{}).
			Where("partner_id = ? AND status = ? AND payout_id IS NULL AND DATE(trade_date) < ?", g.PartnerID, "pending", today).
			Updates(map[string]interface{}{
				"payout_id": payout.ID,
				"status":    "approved",
			}).Error; err != nil {
			return err
		}

		var commissions []database.Commission
		if err := tx.Where("payout_id = ?", payout.ID).Find(&commissions).Error; err != nil {
			return err
		}

		for _, c := range commissions {
			item := database.PayoutItem{
				PayoutID:     payout.ID,
				CommissionID: c.ID,
				Amount:       c.CommissionAmount,
			}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
