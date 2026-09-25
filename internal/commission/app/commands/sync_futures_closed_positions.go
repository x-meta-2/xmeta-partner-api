package commands

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/commission/domain"
	"xmeta-partner/internal/commission/port"
	"xmeta-partner/structs"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

const (
	defaultFuturesClosedPositionSyncLimit = 1000
	maxFuturesClosedPositionSyncLimit     = 10000
	futuresPositionIdentitySQL            = "COALESCE(NULLIF(futures_closed_positions.sk, ''), NULLIF(futures_closed_positions.id, ''), NULLIF(futures_closed_positions.order_id, ''), NULLIF(futures_closed_positions.position_id, ''))"
)

type SyncFuturesClosedPositionsHandler struct {
	DB   *gorm.DB
	Repo port.TradeEventRepo
}

type FuturesClosedPositionSyncResult struct {
	Total   int                `json:"total"`
	Success int                `json:"success"`
	Skipped int                `json:"skipped"`
	Failed  int                `json:"failed"`
	Errors  []FuturesSyncError `json:"errors,omitempty"`
}

type FuturesSyncError struct {
	PositionID string `json:"positionId"`
	UserID     string `json:"userId"`
	Message    string `json:"message"`
}

func (h *SyncFuturesClosedPositionsHandler) Handle(params structs.FuturesClosedPositionSyncParams) (FuturesClosedPositionSyncResult, error) {
	startedAt, endedAt, err := parseSyncRange(params.StartedAt, params.EndedAt)
	if err != nil {
		return FuturesClosedPositionSyncResult{}, err
	}

	feeRate := viper.GetFloat64("FUTURES_FEE_RATE")
	if feeRate <= 0 {
		return FuturesClosedPositionSyncResult{}, domain.ErrFuturesFeeRate
	}

	limit := params.Limit
	if limit <= 0 {
		limit = defaultFuturesClosedPositionSyncLimit
	}
	if limit > maxFuturesClosedPositionSyncLimit {
		limit = maxFuturesClosedPositionSyncLimit
	}

	var positions []database.FuturesClosedPosition
	if err := h.DB.
		Joins("JOIN users ON users.id = futures_closed_positions.user_id AND users.kyc_level >= 1").
		Joins("JOIN referrals ON referrals.referred_user_id = futures_closed_positions.user_id AND referrals.started_at <= futures_closed_positions.closed_at AND (referrals.ended_at IS NULL OR referrals.ended_at > futures_closed_positions.closed_at)").
		Joins("LEFT JOIN commissions ON commissions.position_id = "+futuresPositionIdentitySQL).
		Where("futures_closed_positions.closed_at >= ? AND futures_closed_positions.closed_at < ?", startedAt, endedAt).
		Where(futuresPositionIdentitySQL + " IS NOT NULL").
		Where("futures_closed_positions.user_id <> ''").
		Where("futures_closed_positions.close_notional > 0").
		Where("commissions.position_id IS NULL").
		Order("futures_closed_positions.closed_at asc").
		Limit(limit).
		Find(&positions).Error; err != nil {
		return FuturesClosedPositionSyncResult{}, err
	}

	result := FuturesClosedPositionSyncResult{Total: len(positions)}
	for _, position := range positions {
		created, reason, err := h.processPosition(position, feeRate)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, FuturesSyncError{
				PositionID: commissionPositionID(position),
				UserID:     position.UserID,
				Message:    err.Error(),
			})
			continue
		}
		if !created {
			result.Skipped++
			if reason != "" {
				result.Errors = append(result.Errors, FuturesSyncError{
					PositionID: commissionPositionID(position),
					UserID:     position.UserID,
					Message:    reason,
				})
			}
			continue
		}
		result.Success++
	}

	return result, nil
}

func (h *SyncFuturesClosedPositionsHandler) processPosition(position database.FuturesClosedPosition, feeRate float64) (bool, string, error) {
	positionID := commissionPositionID(position)
	if positionID == "" {
		return false, "missing position identity", nil
	}
	if strings.TrimSpace(position.UserID) == "" {
		return false, "missing userId", nil
	}
	if position.ClosedAt.IsZero() {
		return false, "missing closedAt", nil
	}
	if position.CloseNotional <= 0 {
		return false, "zero closeNotional", nil
	}

	tradeFee := truncate4(position.CloseNotional * feeRate)
	if tradeFee <= 0 {
		return false, "zero derived fee", nil
	}

	kycVerified, err := h.Repo.IsUserKycVerified(position.UserID)
	if err != nil {
		return false, "", err
	}
	if !kycVerified {
		return false, "user not kyc verified", nil
	}

	exists, err := h.Repo.ExistsByPositionID(positionID)
	if err != nil {
		return false, "", err
	}
	if exists {
		return false, "duplicate position", nil
	}

	referral, err := h.Repo.FindActiveReferral(position.UserID, position.ClosedAt)
	if err != nil {
		if errors.Is(err, domain.ErrNoActiveReferral) {
			return false, "no active referral", nil
		}
		return false, "", err
	}

	partner, err := h.Repo.FindPartnerWithTier(referral.PartnerID)
	if err != nil {
		return false, "", domain.ErrPartnerNotFound
	}
	if partner.UserID == position.UserID {
		return false, "self-trade", nil
	}
	if partner.Tier == nil {
		return false, "", domain.ErrNoTierAssigned
	}

	commissionRate := partner.Tier.CommissionRate
	rebateAmount := truncate4(tradeFee * commissionRate)
	commission := database.Commission{
		PartnerID:        partner.ID,
		ReferredUserID:   position.UserID,
		PositionID:       positionID,
		MarketID:         futuresMarketID(position),
		Asset:            "USDT",
		CommissionAmount: tradeFee,
		VolumeUSD:        position.CloseNotional,
		CommissionRate:   commissionRate,
		RebateAmount:     rebateAmount,
		TierID:           &partner.TierID,
		Status:           database.CommissionStatusPending,
		TradeDate:        position.ClosedAt,
	}

	err = h.Repo.RunInTx(func(txRepo port.TradeEventRepo) error {
		if err := txRepo.CreateCommission(&commission); err != nil {
			return err
		}
		if err := txRepo.IncrementPartnerEarnings(partner.ID, rebateAmount); err != nil {
			return err
		}
		if referral.FirstTradeAt == nil {
			if err := txRepo.ActivateReferral(referral.ID, position.ClosedAt); err != nil {
				return err
			}
		}
		return maybeUpgradeTierForTrade(txRepo, partner, position.ClosedAt)
	})
	if err != nil {
		if errors.Is(err, domain.ErrDuplicatePosition) {
			return false, "duplicate position", nil
		}
		return false, "", err
	}

	return true, "", nil
}

func parseSyncRange(startValue, endValue string) (time.Time, time.Time, error) {
	startedAt, startDateOnly, err := parseSyncTime(startValue)
	if err != nil {
		return time.Time{}, time.Time{}, domain.ErrInvalidSyncRange
	}
	endedAt, endDateOnly, err := parseSyncTime(endValue)
	if err != nil {
		return time.Time{}, time.Time{}, domain.ErrInvalidSyncRange
	}
	if endDateOnly {
		endedAt = endedAt.AddDate(0, 0, 1)
	}
	if startDateOnly && !endDateOnly && !startedAt.Before(endedAt) {
		return time.Time{}, time.Time{}, domain.ErrInvalidSyncRange
	}
	if !startedAt.Before(endedAt) {
		return time.Time{}, time.Time{}, domain.ErrInvalidSyncRange
	}
	return startedAt, endedAt, nil
}

func parseSyncTime(value string) (time.Time, bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false, fmt.Errorf("empty time")
	}
	if len(value) == len("2006-01-02") {
		location, err := time.LoadLocation("Asia/Ulaanbaatar")
		if err != nil {
			return time.Time{}, true, err
		}
		t, err := time.ParseInLocation("2006-01-02", value, location)
		return t, true, err
	}
	t, err := time.Parse(time.RFC3339Nano, value)
	return t, false, err
}

func commissionPositionID(position database.FuturesClosedPosition) string {
	for _, value := range []string{position.SK, position.ID, position.OrderID, position.PositionID} {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func futuresMarketID(position database.FuturesClosedPosition) string {
	if strings.TrimSpace(position.MarketID) != "" {
		return position.MarketID
	}
	return position.Symbol
}

func truncate4(value float64) float64 {
	return math.Floor(value*1e4) / 1e4
}
