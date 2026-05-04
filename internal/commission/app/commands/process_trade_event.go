package commands

import (
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/commission/domain"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type ProcessTradeEventHandler struct {
	DB *gorm.DB
}

func (h *ProcessTradeEventHandler) Handle(params structs.TradeEventParams) error {
	tradeDate := time.Now()
	if params.TradeTimestamp > 0 {
		tradeDate = time.Unix(params.TradeTimestamp, 0)
	}

	var referral database.Referral
	err := h.DB.
		Where("referred_user_id = ? AND started_at <= ? AND (ended_at IS NULL OR ended_at > ?)", params.UserID, tradeDate, tradeDate).
		First(&referral).Error
	if err != nil {
		return nil
	}

	return h.DB.Transaction(func(tx *gorm.DB) error {
		var partner database.Partner
		if err := tx.Preload("Tier").Where("id = ? AND status = ?", referral.PartnerID, database.PartnerStatusActive).First(&partner).Error; err != nil {
			return domain.ErrPartnerNotFound
		}

		if partner.Tier == nil {
			return domain.ErrNoTierAssigned
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
			Status:           database.CommissionStatusPending,
			TradeDate:        tradeDate,
		}
		if err := tx.Create(&commission).Error; err != nil {
			return err
		}

		if err := tx.Model(&database.Partner{}).
			Where("id = ?", partner.ID).
			Update("total_earnings", gorm.Expr("total_earnings + ?", commissionAmount)).Error; err != nil {
			return err
		}

		if referral.FirstTradeAt == nil {
			if err := tx.Model(&database.Referral{}).
				Where("id = ?", referral.ID).
				Updates(map[string]interface{}{
					"first_trade_at": &tradeDate,
					"status":         database.ReferralStatusActive,
				}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
