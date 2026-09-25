package commands

import (
	"errors"
	"math"
	"strconv"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/commission/domain"
	"xmeta-partner/internal/commission/port"
	"xmeta-partner/structs"
)

func truncate4(v float64) float64 {
	return math.Floor(v*1e4) / 1e4
}

type TradeEventResult struct {
	Created bool
	Skipped bool
	Reason  string
}

type ProcessTradeEventHandler struct {
	Repo port.TradeEventRepo
}

func (h *ProcessTradeEventHandler) Handle(params structs.TradeEventParams) (TradeEventResult, error) {
	tradeFee, err := strconv.ParseFloat(params.CommissionAmount, 64)
	if err != nil || tradeFee <= 0 {
		return TradeEventResult{Skipped: true, Reason: "zero or invalid fee"}, nil
	}
	tradeFee = truncate4(tradeFee)

	tradeDate := time.Now()
	if params.CreatedAt != "" {
		t, err := time.Parse(time.RFC3339Nano, params.CreatedAt)
		if err != nil {
			return TradeEventResult{}, domain.ErrInvalidTradeDate
		}
		tradeDate = t
	}

	kycVerified, err := h.Repo.IsUserKycVerified(params.UserID)
	if err != nil {
		return TradeEventResult{}, err
	}
	if !kycVerified {
		return TradeEventResult{Skipped: true, Reason: "user not kyc verified"}, nil
	}

	exists, err := h.Repo.ExistsByPositionID(params.PositionID)
	if err != nil {
		return TradeEventResult{}, err
	}
	if exists {
		return TradeEventResult{Skipped: true, Reason: "duplicate position"}, nil
	}

	referral, err := h.Repo.FindActiveReferral(params.UserID, tradeDate)
	if err != nil {
		if errors.Is(err, domain.ErrNoActiveReferral) {
			return TradeEventResult{Skipped: true, Reason: "no active referral"}, nil
		}
		return TradeEventResult{}, err
	}

	partner, err := h.Repo.FindActivePartnerWithTier(referral.PartnerID)
	if err != nil {
		return TradeEventResult{}, domain.ErrPartnerNotFound
	}

	if partner.UserID == params.UserID {
		return TradeEventResult{Skipped: true, Reason: "self-trade"}, nil
	}

	if partner.Tier == nil {
		return TradeEventResult{}, domain.ErrNoTierAssigned
	}

	commissionRate := partner.Tier.CommissionRate
	rebateAmount := truncate4(tradeFee * commissionRate)

	var volumeUSD float64
	if v, err := strconv.ParseFloat(params.VolumeInUSD, 64); err == nil {
		volumeUSD = v
	}

	commission := database.Commission{
		PartnerID:        partner.ID,
		ReferredUserID:   params.UserID,
		PositionID:       params.PositionID,
		MarketID:         params.MarketID,
		Asset:            params.CommissionAsset,
		CommissionAmount: tradeFee,
		VolumeUSD:        volumeUSD,
		CommissionRate:   commissionRate,
		RebateAmount:     rebateAmount,
		TierID:           &partner.TierID,
		Status:           database.CommissionStatusPending,
		TradeDate:        tradeDate,
	}

	err = h.Repo.RunInTx(func(txRepo port.TradeEventRepo) error {
		if err := txRepo.CreateCommission(&commission); err != nil {
			return err
		}
		if err := txRepo.IncrementPartnerEarnings(partner.ID, rebateAmount); err != nil {
			return err
		}
		if referral.FirstTradeAt == nil {
			if err := txRepo.ActivateReferral(referral.ID, tradeDate); err != nil {
				return err
			}
		}
		return maybeUpgradeTierForTrade(txRepo, partner, tradeDate)
	})
	if err != nil {
		if errors.Is(err, domain.ErrDuplicatePosition) {
			return TradeEventResult{Skipped: true, Reason: "duplicate position"}, nil
		}
		return TradeEventResult{}, err
	}

	return TradeEventResult{Created: true}, nil
}
