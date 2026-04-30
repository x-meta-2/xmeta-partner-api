package services

import (
	"fmt"
	"log"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type CommissionEngineService struct {
	BaseService
}

// ProcessTradeEvent calculates and records commission for a trade event.
//
// Commission is computed from the partner's tier rate (futures-only — there
// is no per-trade-type override layer). Steps, all in a GORM transaction:
//  1. Find referral by referred_user_id
//  2. Load partner with tier
//  3. commission = trade_fee * tier.commission_rate
//  4. Insert commission record
//  5. Atomically bump partner.total_earnings
//  6. Upsert partner_daily_stats
//  7. If partner has a parent (sub-affiliate), create override commission
func (s *CommissionEngineService) ProcessTradeEvent(params structs.TradeEventParams) error {
	tradeDate := time.Now()
	if params.TradeTimestamp > 0 {
		tradeDate = time.Unix(params.TradeTimestamp, 0)
	}

	// Find which partner the user was attached to *at trade time*. Users
	// can switch partners (close one referral row, open another), so a
	// trade made yesterday must attribute to yesterday's partner — not
	// whoever they happen to be linked to right now.
	var referral database.Referral
	if err := s.DB.Where(
		"referred_user_id = ? AND started_at <= ? AND (ended_at IS NULL OR ended_at > ?)",
		params.UserID, tradeDate, tradeDate,
	).First(&referral).Error; err != nil {
		// User had no active partner at the trade timestamp — skip.
		return nil
	}

	return s.DB.Transaction(func(tx *gorm.DB) error {
		// Get the partner and their tier
		var partner database.Partner
		if err := tx.Preload("Tier").Where("id = ? AND status = ?", referral.PartnerID, "active").First(&partner).Error; err != nil {
			return fmt.Errorf("partner not found or inactive: %s", referral.PartnerID)
		}

		if partner.Tier == nil {
			return fmt.Errorf("partner %s has no tier assigned", partner.ID)
		}

		commissionRate := partner.Tier.CommissionRate
		commissionAmount := params.TradeFee * commissionRate

		// Create commission record
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

		// Atomically increment partner total earnings
		if err := tx.Model(&database.Partner{}).Where("id = ?", partner.ID).
			UpdateColumn("total_earnings", gorm.Expr("total_earnings + ?", commissionAmount)).Error; err != nil {
			return err
		}

		// Upsert partner_daily_stats. Raw INSERT bypasses GORM auto
		// timestamps, so set created_at / updated_at explicitly.
		today := tradeDate.Format("2006-01-02")
		now := time.Now()
		if err := tx.Exec(
			`INSERT INTO partner_daily_stats (id, created_at, updated_at, partner_id, date, trade_volume, commissions)
			 VALUES (gen_random_uuid(), ?, ?, ?, ?, ?, ?)
			 ON CONFLICT (partner_id, date)
			 DO UPDATE SET trade_volume = partner_daily_stats.trade_volume + EXCLUDED.trade_volume,
			               commissions = partner_daily_stats.commissions + EXCLUDED.commissions,
			               updated_at = EXCLUDED.updated_at`,
			now, now, partner.ID, today, params.TradeAmount, commissionAmount,
		).Error; err != nil {
			return err
		}

		// Handle sub-affiliate override commission. We don't want a parent-
		// chain failure to roll back the (already-correct) child trade
		// commission, but silently swallowing the error loses observability.
		if partner.ParentID != nil {
			if err := s.processOverrideCommission(tx, *partner.ParentID, params, commission); err != nil {
				log.Printf("[commission] override failed for parent=%s child=%s trade=%s: %v",
					*partner.ParentID, partner.ID, params.TradeID, err)
			}
		}

		// First-trade transition: mark referral active so dashboards/stats
		// reflect the user as an "active client".
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

// processOverrideCommission creates override commission for parent partner within a transaction
func (s *CommissionEngineService) processOverrideCommission(tx *gorm.DB, parentID string, params structs.TradeEventParams, childCommission database.Commission) error {
	// Verify parent partner is active
	var parentPartner database.Partner
	if err := tx.Where("id = ? AND status = ?", parentID, "active").First(&parentPartner).Error; err != nil {
		return nil // Parent not active, skip
	}

	// Get the sub-affiliate invite to find override rate
	var invite database.SubAffiliateInvite
	if err := tx.Where("sub_partner_id = ? AND status = ?", childCommission.PartnerID, "accepted").
		First(&invite).Error; err != nil {
		return nil // No accepted invite found, skip
	}

	overrideAmount := childCommission.CommissionAmount * invite.OverrideRate

	override := database.Commission{
		PartnerID:         parentID,
		ReferredUserID:    params.UserID,
		TradeID:           params.TradeID,
		TradeAmount:       params.TradeAmount,
		CommissionRate:    invite.OverrideRate,
		CommissionAmount:  overrideAmount,
		IsOverride:        true,
		OverridePartnerID: &childCommission.PartnerID,
		Status:            "pending",
		TradeDate:         childCommission.TradeDate,
	}

	if err := tx.Create(&override).Error; err != nil {
		return err
	}

	// Atomically update parent total earnings
	if err := tx.Model(&database.Partner{}).Where("id = ?", parentID).
		UpdateColumn("total_earnings", gorm.Expr("total_earnings + ?", overrideAmount)).Error; err != nil {
		return err
	}

	// Upsert parent daily stats. Mirror the child's columns (trade_volume +
	// commissions) so the parent dashboard shows the same totals that drove
	// their override earnings.
	today := childCommission.TradeDate.Format("2006-01-02")
	now := time.Now()
	return tx.Exec(
		`INSERT INTO partner_daily_stats (id, created_at, updated_at, partner_id, date, trade_volume, commissions)
		 VALUES (gen_random_uuid(), ?, ?, ?, ?, ?, ?)
		 ON CONFLICT (partner_id, date)
		 DO UPDATE SET trade_volume = partner_daily_stats.trade_volume + EXCLUDED.trade_volume,
		               commissions = partner_daily_stats.commissions + EXCLUDED.commissions,
		               updated_at = EXCLUDED.updated_at`,
		now, now,
		parentID, today, params.TradeAmount, overrideAmount,
	).Error
}
