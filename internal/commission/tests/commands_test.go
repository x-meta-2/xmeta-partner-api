package tests

import (
	"testing"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/commission/app/commands"
	"xmeta-partner/internal/commission/domain"
	"xmeta-partner/structs"

	"github.com/stretchr/testify/assert"
)

func newTestPartner() *database.Partner {
	return &database.Partner{
		Base:   database.Base{ID: "partner-1"},
		UserID: "partner-user-1",
		TierID: "tier-1",
		Status: database.PartnerStatusActive,
		Tier: &database.PartnerTier{
			Base:           database.Base{ID: "tier-1"},
			CommissionRate: 0.30,
		},
	}
}

func newTestReferral() *database.Referral {
	now := time.Now()
	return &database.Referral{
		Base:           database.Base{ID: "ref-1"},
		PartnerID:      "partner-1",
		ReferredUserID: "user-1",
		StartedAt:      now.Add(-24 * time.Hour),
	}
}

func baseRepo() *TradeEventRepo {
	return &TradeEventRepo{
		ExistsByPositionIDFn:        func(string) (bool, error) { return false, nil },
		IsUserKycVerifiedFn:         func(string) (bool, error) { return true, nil },
		FindActiveReferralFn:        func(string, time.Time) (*database.Referral, error) { return newTestReferral(), nil },
		FindActivePartnerWithTierFn: func(string) (*database.Partner, error) { return newTestPartner(), nil },
		CreateCommissionFn:          func(*database.Commission) error { return nil },
		IncrementPartnerEarningsFn:  func(string, float64) error { return nil },
		ActivateReferralFn:          func(string, time.Time) error { return nil },
	}
}

func TestProcessTradeEvent_Success(t *testing.T) {
	var created *database.Commission
	var earningsAdded float64

	repo := baseRepo()
	repo.CreateCommissionFn = func(c *database.Commission) error {
		created = c
		return nil
	}
	repo.IncrementPartnerEarningsFn = func(_ string, amount float64) error {
		earningsAdded = amount
		return nil
	}

	handler := commands.ProcessTradeEventHandler{Repo: repo}
	res, err := handler.Handle(structs.TradeEventParams{
		UserID:           "user-1",
		PositionID:       "pos-123",
		MarketID:         "perp.eth_usdt",
		CommissionAsset:  "USDT",
		CommissionAmount: "10.0",
		VolumeInUSD:      "5000.0",
		CreatedAt:        "2026-05-05T10:00:00Z",
	})

	assert.NoError(t, err)
	assert.True(t, res.Created)
	assert.False(t, res.Skipped)
	assert.NotNil(t, created)
	assert.Equal(t, "partner-1", created.PartnerID)
	assert.Equal(t, "user-1", created.ReferredUserID)
	assert.Equal(t, "pos-123", created.PositionID)
	assert.Equal(t, "perp.eth_usdt", created.MarketID)
	assert.Equal(t, "USDT", created.Asset)
	assert.Equal(t, 10.0, created.CommissionAmount)
	assert.Equal(t, 5000.0, created.VolumeUSD)
	assert.Equal(t, 0.30, created.CommissionRate)
	assert.Equal(t, 3.0, created.RebateAmount)
	assert.Equal(t, database.CommissionStatusPending, created.Status)
	assert.Equal(t, 3.0, earningsAdded)
}

func TestProcessTradeEvent_ZeroFee_Skips(t *testing.T) {
	handler := commands.ProcessTradeEventHandler{Repo: baseRepo()}
	res, err := handler.Handle(structs.TradeEventParams{
		UserID:           "user-1",
		PositionID:       "pos-1",
		CommissionAmount: "0",
	})

	assert.NoError(t, err)
	assert.True(t, res.Skipped)
	assert.Equal(t, "zero or invalid fee", res.Reason)
}

func TestProcessTradeEvent_InvalidFee_Skips(t *testing.T) {
	handler := commands.ProcessTradeEventHandler{Repo: baseRepo()}
	res, err := handler.Handle(structs.TradeEventParams{
		UserID:           "user-1",
		PositionID:       "pos-1",
		CommissionAmount: "not_a_number",
	})

	assert.NoError(t, err)
	assert.True(t, res.Skipped)
}

func TestProcessTradeEvent_DuplicatePosition_Skips(t *testing.T) {
	repo := baseRepo()
	repo.ExistsByPositionIDFn = func(string) (bool, error) { return true, nil }

	handler := commands.ProcessTradeEventHandler{Repo: repo}
	res, err := handler.Handle(structs.TradeEventParams{
		UserID:           "user-1",
		PositionID:       "pos-duplicate",
		CommissionAmount: "10.0",
	})

	assert.NoError(t, err)
	assert.True(t, res.Skipped)
	assert.Equal(t, "duplicate position", res.Reason)
}

func TestProcessTradeEvent_NoReferral_Skips(t *testing.T) {
	repo := baseRepo()
	repo.FindActiveReferralFn = func(string, time.Time) (*database.Referral, error) {
		return nil, domain.ErrNoActiveReferral
	}

	handler := commands.ProcessTradeEventHandler{Repo: repo}
	res, err := handler.Handle(structs.TradeEventParams{
		UserID:           "user-no-referral",
		PositionID:       "pos-2",
		CommissionAmount: "10.0",
	})

	assert.NoError(t, err)
	assert.True(t, res.Skipped)
	assert.Equal(t, "no active referral", res.Reason)
}

func TestProcessTradeEvent_SelfTrade_Skips(t *testing.T) {
	var created *database.Commission

	repo := baseRepo()
	partner := newTestPartner()
	partner.UserID = "user-self"
	repo.FindActivePartnerWithTierFn = func(string) (*database.Partner, error) { return partner, nil }
	repo.CreateCommissionFn = func(c *database.Commission) error {
		created = c
		return nil
	}

	handler := commands.ProcessTradeEventHandler{Repo: repo}
	res, err := handler.Handle(structs.TradeEventParams{
		UserID:           "user-self",
		PositionID:       "pos-3",
		CommissionAmount: "10.0",
	})

	assert.NoError(t, err)
	assert.True(t, res.Skipped)
	assert.Equal(t, "self-trade", res.Reason)
	assert.Nil(t, created)
}

func TestProcessTradeEvent_PartnerNotFound_Error(t *testing.T) {
	repo := baseRepo()
	repo.FindActivePartnerWithTierFn = func(string) (*database.Partner, error) {
		return nil, domain.ErrPartnerNotFound
	}

	handler := commands.ProcessTradeEventHandler{Repo: repo}
	_, err := handler.Handle(structs.TradeEventParams{
		UserID:           "user-1",
		PositionID:       "pos-4",
		CommissionAmount: "10.0",
	})

	assert.ErrorIs(t, err, domain.ErrPartnerNotFound)
}

func TestProcessTradeEvent_NoTier_Error(t *testing.T) {
	repo := baseRepo()
	partner := newTestPartner()
	partner.Tier = nil
	repo.FindActivePartnerWithTierFn = func(string) (*database.Partner, error) { return partner, nil }

	handler := commands.ProcessTradeEventHandler{Repo: repo}
	_, err := handler.Handle(structs.TradeEventParams{
		UserID:           "user-1",
		PositionID:       "pos-5",
		CommissionAmount: "10.0",
	})

	assert.ErrorIs(t, err, domain.ErrNoTierAssigned)
}

func TestProcessTradeEvent_CreateFails_Error(t *testing.T) {
	repo := baseRepo()
	repo.CreateCommissionFn = func(*database.Commission) error { return errDB }

	handler := commands.ProcessTradeEventHandler{Repo: repo}
	_, err := handler.Handle(structs.TradeEventParams{
		UserID:           "user-1",
		PositionID:       "pos-6",
		CommissionAmount: "10.0",
	})

	assert.ErrorIs(t, err, errDB)
}

func TestProcessTradeEvent_ActivatesReferral(t *testing.T) {
	var activated bool

	referral := newTestReferral()
	referral.FirstTradeAt = nil

	repo := baseRepo()
	repo.FindActiveReferralFn = func(string, time.Time) (*database.Referral, error) { return referral, nil }
	repo.ActivateReferralFn = func(string, time.Time) error {
		activated = true
		return nil
	}

	handler := commands.ProcessTradeEventHandler{Repo: repo}
	res, _ := handler.Handle(structs.TradeEventParams{
		UserID:           "user-1",
		PositionID:       "pos-7",
		CommissionAmount: "5.0",
	})

	assert.True(t, res.Created)
	assert.True(t, activated)
}

func TestProcessTradeEvent_AlreadyActivated_NoReactivate(t *testing.T) {
	var activated bool
	now := time.Now()

	referral := newTestReferral()
	referral.FirstTradeAt = &now

	repo := baseRepo()
	repo.FindActiveReferralFn = func(string, time.Time) (*database.Referral, error) { return referral, nil }
	repo.ActivateReferralFn = func(string, time.Time) error {
		activated = true
		return nil
	}

	handler := commands.ProcessTradeEventHandler{Repo: repo}
	res, _ := handler.Handle(structs.TradeEventParams{
		UserID:           "user-1",
		PositionID:       "pos-8",
		CommissionAmount: "5.0",
	})

	assert.True(t, res.Created)
	assert.False(t, activated)
}

func TestProcessTradeEvent_NoKyc_Skips(t *testing.T) {
	repo := baseRepo()
	repo.IsUserKycVerifiedFn = func(string) (bool, error) { return false, nil }

	handler := commands.ProcessTradeEventHandler{Repo: repo}
	res, err := handler.Handle(structs.TradeEventParams{
		UserID:           "user-no-kyc",
		PositionID:       "pos-9",
		CommissionAmount: "10.0",
	})

	assert.NoError(t, err)
	assert.True(t, res.Skipped)
	assert.Equal(t, "user not kyc verified", res.Reason)
}

func TestProcessTradeEvent_UserNotFound_Skips(t *testing.T) {
	repo := baseRepo()
	repo.IsUserKycVerifiedFn = func(string) (bool, error) { return false, errDB }

	handler := commands.ProcessTradeEventHandler{Repo: repo}
	res, err := handler.Handle(structs.TradeEventParams{
		UserID:           "user-nonexistent",
		PositionID:       "pos-10",
		CommissionAmount: "10.0",
	})

	assert.NoError(t, err)
	assert.True(t, res.Skipped)
	assert.Equal(t, "user not found", res.Reason)
}
