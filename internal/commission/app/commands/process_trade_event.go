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

func truncate8(v float64) float64 {
	return math.Round(v*1e8) / 1e8
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
	rebateAmount := truncate8(tradeFee * commissionRate)

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
		return h.maybeUpgradeTier(txRepo, partner)
	})
	if err != nil {
		if errors.Is(err, domain.ErrDuplicatePosition) {
			return TradeEventResult{Skipped: true, Reason: "duplicate position"}, nil
		}
		return TradeEventResult{}, err
	}

	return TradeEventResult{Created: true}, nil
}

func (h *ProcessTradeEventHandler) maybeUpgradeTier(txRepo port.TradeEventRepo, partner *database.Partner) error {
	totalVolume, err := txRepo.GetPartnerTotalVolume(partner.ID)
	if err != nil {
		return err
	}

	activeClients, err := txRepo.GetPartnerActiveClients(partner.ID)
	if err != nil {
		return err
	}

	tiers, err := txRepo.FindAllTiersAsc()
	if err != nil {
		return err
	}

	var bestTier *database.PartnerTier
	for i := range tiers {
		t := &tiers[i]
		volumeOK := t.MinVolume == 0 || totalVolume >= t.MinVolume
		clientsOK := t.MinActiveClients == 0 || activeClients >= int64(t.MinActiveClients)
		if volumeOK && clientsOK {
			bestTier = t
		}
	}

	if bestTier != nil && bestTier.Level > partner.Tier.Level {
		return txRepo.UpgradePartnerTier(partner.ID, bestTier.ID, bestTier.Level)
	}

	return nil
}
