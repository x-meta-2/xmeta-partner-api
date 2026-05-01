package services

import (
	"fmt"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type CommissionEngineService struct {
	BaseService
}

// ProcessTradeEvent computes and records commission for a trade event.
// Attribution is time-bounded: a trade attributes to whichever partner the
// user was linked to at the trade timestamp, not whoever they're linked to now.
func (s *CommissionEngineService) ProcessTradeEvent(params structs.TradeEventParams) error {
	tradeDate := time.Now()
	if params.TradeTimestamp > 0 {
		tradeDate = time.Unix(params.TradeTimestamp, 0)
	}

	var referral database.Referral
	if err := s.DB.Where(
		"referred_user_id = ? AND started_at <= ? AND (ended_at IS NULL OR ended_at > ?)",
		params.UserID, tradeDate, tradeDate,
	).First(&referral).Error; err != nil {
		return nil
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		var partner database.Partner
		if err := tx.Preload("Tier").Where("id = ? AND status = ?", referral.PartnerID, "active").First(&partner).Error; err != nil {
			return fmt.Errorf("partner not found or inactive: %s", referral.PartnerID)
		}

		if partner.Tier == nil {
			return fmt.Errorf("partner %s has no tier assigned", partner.ID)
		}

		commissionRate := partner.Tier.CommissionRate
		commissionAmount := params.TradeFee * commissionRate

		commission := database.Commission{
			PartnerID:        partner.ID,
			ReferredUserID:   params.UserID,
			TradeID:          params.TradeID,
			TradeAmount:      params.TradeAmount,
			CommissionRate:   commissionRate,
			CommissionAmount: commissionAmount,
			TierID:           &partner.TierID,
			Status:           "pending",
			TradeDate:        tradeDate,
		}

		if err := tx.Create(&commission).Error; err != nil {
			return err
		}

		if err := tx.Model(&database.Partner{}).Where("id = ?", partner.ID).
			UpdateColumn("total_earnings", gorm.Expr("total_earnings + ?", commissionAmount)).Error; err != nil {
			return err
		}

		if referral.FirstTradeAt == nil {
			if err := tx.Model(&referral).Updates(map[string]interface{}{
				"first_trade_at": &tradeDate,
				"status":         database.ReferralStatusActive,
			}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

