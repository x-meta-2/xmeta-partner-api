package tests

import (
	"testing"

	"xmeta-partner/database"
	"xmeta-partner/internal/partner/app/commands"
	"xmeta-partner/internal/partner/domain"
	"xmeta-partner/internal/partner/port"
	"xmeta-partner/structs"

	"github.com/stretchr/testify/assert"
)

func TestRejectApplication_Success(t *testing.T) {
	var saved *database.PartnerApplication

	apps := &ApplicationRepo{
		FindPendingByIDFn: func(id string) (*database.PartnerApplication, error) {
			return &database.PartnerApplication{
				Base:   database.Base{ID: id},
				Status: database.ApplicationStatusPending,
			}, nil
		},
		SaveFn: func(app *database.PartnerApplication) error {
			saved = app
			return nil
		},
	}

	handler := commands.RejectApplicationHandler{Apps: apps}
	result, err := handler.Handle("app-1", "admin-1", "does not meet criteria")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, database.ApplicationStatusRejected, result.Status)
	assert.Equal(t, "does not meet criteria", result.RejectionReason)
	assert.NotNil(t, result.ReviewedBy)
	assert.Equal(t, "admin-1", *result.ReviewedBy)
	assert.NotNil(t, result.ReviewedAt)
	assert.Equal(t, saved, result)
}

func TestRejectApplication_NotFound(t *testing.T) {
	apps := &ApplicationRepo{
		FindPendingByIDFn: func(id string) (*database.PartnerApplication, error) {
			return nil, domain.ErrApplicationNotFound
		},
	}

	handler := commands.RejectApplicationHandler{Apps: apps}
	result, err := handler.Handle("missing", "admin-1", "reason")

	assert.ErrorIs(t, err, domain.ErrApplicationNotFound)
	assert.Nil(t, result)
}

func TestRejectApplication_SaveFails(t *testing.T) {
	apps := &ApplicationRepo{
		FindPendingByIDFn: func(id string) (*database.PartnerApplication, error) {
			return &database.PartnerApplication{
				Base:   database.Base{ID: id},
				Status: database.ApplicationStatusPending,
			}, nil
		},
		SaveFn: func(app *database.PartnerApplication) error {
			return errDB
		},
	}

	handler := commands.RejectApplicationHandler{Apps: apps}
	result, err := handler.Handle("app-1", "admin-1", "reason")

	assert.ErrorIs(t, err, errDB)
	assert.Nil(t, result)
}

// ── DeleteTier ──

func TestDeleteTier_DefaultTier_Rejected(t *testing.T) {
	tiers := &TierRepo{
		FindByIDFn: func(id string) (*database.PartnerTier, error) {
			return &database.PartnerTier{Base: database.Base{ID: id}, IsDefault: true}, nil
		},
	}

	handler := commands.DeleteTierHandler{Tiers: tiers}
	err := handler.Handle("tier-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete the default tier")
}

func TestDeleteTier_WithPartners_Rejected(t *testing.T) {
	tiers := &TierRepo{
		FindByIDFn: func(id string) (*database.PartnerTier, error) {
			return &database.PartnerTier{Base: database.Base{ID: id}, IsDefault: false}, nil
		},
		CountPartnersByTierFn: func(tierID string) (int64, error) { return 5, nil },
	}

	handler := commands.DeleteTierHandler{Tiers: tiers}
	err := handler.Handle("tier-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "5 partners are using it")
}

func TestDeleteTier_Success(t *testing.T) {
	deleted := false
	tiers := &TierRepo{
		FindByIDFn: func(id string) (*database.PartnerTier, error) {
			return &database.PartnerTier{Base: database.Base{ID: id}, IsDefault: false}, nil
		},
		CountPartnersByTierFn: func(tierID string) (int64, error) { return 0, nil },
		DeleteFn: func(id string) error {
			deleted = true
			return nil
		},
	}

	handler := commands.DeleteTierHandler{Tiers: tiers}
	err := handler.Handle("tier-1")

	assert.NoError(t, err)
	assert.True(t, deleted)
}

func TestDeleteTier_NotFound(t *testing.T) {
	tiers := &TierRepo{
		FindByIDFn: func(id string) (*database.PartnerTier, error) { return nil, errDB },
	}

	handler := commands.DeleteTierHandler{Tiers: tiers}
	err := handler.Handle("missing")

	assert.ErrorIs(t, err, errDB)
}

// ── CreateTier ──

func TestCreateTier_DefaultClearsOthers(t *testing.T) {
	cleared := false
	created := false

	tiers := &TierRepo{
		ClearDefaultExceptFn: func(tierID string) error {
			cleared = true
			assert.Equal(t, "", tierID)
			return nil
		},
		CreateFn: func(tier *database.PartnerTier) error {
			created = true
			return nil
		},
	}

	handler := commands.CreateTierHandler{Tiers: tiers}
	_, err := handler.Handle(structs.TierCreateParams{
		Name: "Gold", Level: 2, CommissionRate: 0.35, IsDefault: true, Color: "#ffd700",
	})

	assert.NoError(t, err)
	assert.True(t, cleared)
	assert.True(t, created)
}

func TestCreateTier_ClearDefaultFails_Propagates(t *testing.T) {
	tiers := &TierRepo{
		ClearDefaultExceptFn: func(tierID string) error { return errDB },
	}

	handler := commands.CreateTierHandler{Tiers: tiers}
	_, err := handler.Handle(structs.TierCreateParams{
		Name: "Gold", Level: 2, CommissionRate: 0.35, IsDefault: true,
	})

	assert.ErrorIs(t, err, errDB)
}

func TestCreateTier_NonDefault_NoClear(t *testing.T) {
	created := false
	tiers := &TierRepo{
		CreateFn: func(tier *database.PartnerTier) error {
			created = true
			assert.False(t, tier.IsDefault)
			return nil
		},
	}

	handler := commands.CreateTierHandler{Tiers: tiers}
	_, err := handler.Handle(structs.TierCreateParams{
		Name: "Silver", Level: 1, CommissionRate: 0.20, IsDefault: false,
	})

	assert.NoError(t, err)
	assert.True(t, created)
}

// ── UpdateTier ──

func TestUpdateTier_UnsetDefault_Rejected(t *testing.T) {
	tiers := &TierRepo{
		FindByIDFn: func(id string) (*database.PartnerTier, error) {
			return &database.PartnerTier{Base: database.Base{ID: id}, IsDefault: true}, nil
		},
	}

	handler := commands.UpdateTierHandler{Tiers: tiers}
	isDefault := false
	_, err := handler.Handle("tier-1", structs.TierUpdateParams{IsDefault: &isDefault})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot unset the default tier")
}

func TestUpdateTier_SetDefault_ClearsOthers(t *testing.T) {
	cleared := false
	updated := false

	tiers := &TierRepo{
		FindByIDFn: func(id string) (*database.PartnerTier, error) {
			if !updated {
				return &database.PartnerTier{Base: database.Base{ID: id}, IsDefault: false, Name: "Silver"}, nil
			}
			return &database.PartnerTier{Base: database.Base{ID: id}, IsDefault: true, Name: "Silver"}, nil
		},
		ClearDefaultExceptFn: func(tierID string) error {
			cleared = true
			assert.Equal(t, "tier-2", tierID)
			return nil
		},
		UpdateFn: func(id string, fields map[string]interface{}) error {
			updated = true
			assert.Equal(t, true, fields["is_default"])
			return nil
		},
	}

	handler := commands.UpdateTierHandler{Tiers: tiers}
	isDefault := true
	result, err := handler.Handle("tier-2", structs.TierUpdateParams{IsDefault: &isDefault})

	assert.NoError(t, err)
	assert.True(t, cleared)
	assert.True(t, updated)
	assert.True(t, result.IsDefault)
}

func TestUpdateTier_ClearDefaultFails_Propagates(t *testing.T) {
	tiers := &TierRepo{
		FindByIDFn: func(id string) (*database.PartnerTier, error) {
			return &database.PartnerTier{Base: database.Base{ID: id}, IsDefault: false}, nil
		},
		ClearDefaultExceptFn: func(tierID string) error { return errDB },
	}

	handler := commands.UpdateTierHandler{Tiers: tiers}
	isDefault := true
	_, err := handler.Handle("tier-1", structs.TierUpdateParams{IsDefault: &isDefault})

	assert.ErrorIs(t, err, errDB)
}

func TestUpdateTier_TxUsed_ForDefaultChange(t *testing.T) {
	txUsed := false
	tiers := &TierRepo{
		FindByIDFn: func(id string) (*database.PartnerTier, error) {
			return &database.PartnerTier{Base: database.Base{ID: id}, IsDefault: false}, nil
		},
		ClearDefaultExceptFn: func(tierID string) error { return nil },
		UpdateFn:             func(id string, fields map[string]interface{}) error { return nil },
	}

	wrappedTiers := &txTrackingTierRepo{TierRepo: tiers, txUsed: &txUsed}

	handler := commands.UpdateTierHandler{Tiers: wrappedTiers}
	isDefault := true
	_, err := handler.Handle("tier-1", structs.TierUpdateParams{IsDefault: &isDefault})

	assert.NoError(t, err)
	assert.True(t, txUsed)
}

type txTrackingTierRepo struct {
	*TierRepo
	txUsed *bool
}

func (t *txTrackingTierRepo) RunInTx(fn func(port.TierRepo) error) error {
	*t.txUsed = true
	return fn(t.TierRepo)
}
