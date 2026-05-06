package tests

import (
	"testing"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/commission/app/queries"
	"xmeta-partner/structs"

	"github.com/stretchr/testify/assert"
)

func TestAdminListCommissions_Success(t *testing.T) {
	commissions := []database.Commission{
		{
			Base:             database.Base{ID: "c-1"},
			PartnerID:        "p-1",
			ReferredUserID:   "u-1",
			PositionID:       "pos-1",
			MarketID:         "perp.btc_usdt",
			Asset:            "USDT",
			CommissionAmount: 10.0,
			VolumeUSD:        5000.0,
			CommissionRate:   0.30,
			RebateAmount:     3.0,
			Status:           database.CommissionStatusPending,
			TradeDate:        time.Now(),
		},
	}

	repo := &CommissionRepo{
		AdminListFn: func(params structs.AdminCommissionListParams) ([]database.Commission, int, error) {
			return commissions, 1, nil
		},
	}

	handler := queries.AdminListCommissionsHandler{Commissions: repo}
	result, err := handler.Handle(structs.AdminCommissionListParams{})

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Len(t, result.Items.([]database.Commission), 1)
}

func TestAdminListCommissions_Empty(t *testing.T) {
	repo := &CommissionRepo{
		AdminListFn: func(params structs.AdminCommissionListParams) ([]database.Commission, int, error) {
			return []database.Commission{}, 0, nil
		},
	}

	handler := queries.AdminListCommissionsHandler{Commissions: repo}
	result, err := handler.Handle(structs.AdminCommissionListParams{})

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Total)
}

func TestAdminListCommissions_DBError(t *testing.T) {
	repo := &CommissionRepo{
		AdminListFn: func(params structs.AdminCommissionListParams) ([]database.Commission, int, error) {
			return nil, 0, errDB
		},
	}

	handler := queries.AdminListCommissionsHandler{Commissions: repo}
	result, err := handler.Handle(structs.AdminCommissionListParams{})

	assert.ErrorIs(t, err, errDB)
	assert.Equal(t, 0, result.Total)
}

func TestListCommissions_Success(t *testing.T) {
	commissions := []database.Commission{
		{Base: database.Base{ID: "c-1"}, PartnerID: "p-1", Status: database.CommissionStatusApproved},
		{Base: database.Base{ID: "c-2"}, PartnerID: "p-1", Status: database.CommissionStatusPending},
	}

	repo := &CommissionRepo{
		ListFn: func(partnerID string, params structs.CommissionListParams) ([]database.Commission, int, error) {
			assert.Equal(t, "p-1", partnerID)
			return commissions, 2, nil
		},
	}

	handler := queries.ListCommissionsHandler{Commissions: repo}
	result, err := handler.Handle("p-1", structs.CommissionListParams{})

	assert.NoError(t, err)
	assert.Equal(t, 2, result.Total)
}

func TestCommissionBreakdown_Success(t *testing.T) {
	repo := &CommissionRepo{
		BreakdownFn: func(partnerID string, params structs.CommissionBreakdownParams) (float64, error) {
			return 150.50, nil
		},
	}

	handler := queries.CommissionBreakdownHandler{Commissions: repo}
	result, err := handler.Handle("p-1", structs.CommissionBreakdownParams{})

	assert.NoError(t, err)
	assert.Equal(t, 150.50, result.Futures)
	assert.Equal(t, 150.50, result.Total)
}
