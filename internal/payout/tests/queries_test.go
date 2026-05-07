package tests

import (
	"testing"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/payout/app/queries"
	"xmeta-partner/structs"

	"github.com/stretchr/testify/assert"
)

// ── ListPayouts ──

func TestListPayouts_Success(t *testing.T) {
	payouts := []database.Payout{
		{Base: database.Base{ID: "p-1"}, Amount: 100, Status: database.PayoutStatusPending},
		{Base: database.Base{ID: "p-2"}, Amount: 200, Status: database.PayoutStatusCompleted},
	}

	repo := &PayoutRepo{
		ListFn: func(partnerID string, params structs.PayoutListParams) ([]database.Payout, int, error) {
			assert.Equal(t, "partner-1", partnerID)
			return payouts, 2, nil
		},
	}

	handler := queries.ListPayoutsHandler{Payouts: repo}
	result, err := handler.Handle("partner-1", structs.PayoutListParams{})

	assert.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	assert.Len(t, result.Items, 2)
}

func TestListPayouts_Empty(t *testing.T) {
	repo := &PayoutRepo{
		ListFn: func(partnerID string, params structs.PayoutListParams) ([]database.Payout, int, error) {
			return []database.Payout{}, 0, nil
		},
	}

	handler := queries.ListPayoutsHandler{Payouts: repo}
	result, err := handler.Handle("partner-1", structs.PayoutListParams{})

	assert.NoError(t, err)
	assert.Equal(t, 0, result.Total)
	assert.Empty(t, result.Items)
}

func TestListPayouts_RepoError(t *testing.T) {
	repo := &PayoutRepo{
		ListFn: func(partnerID string, params structs.PayoutListParams) ([]database.Payout, int, error) {
			return nil, 0, errDB
		},
	}

	handler := queries.ListPayoutsHandler{Payouts: repo}
	result, err := handler.Handle("partner-1", structs.PayoutListParams{})

	assert.ErrorIs(t, err, errDB)
	assert.Equal(t, 0, result.Total)
}

func TestListPayouts_PassesParams(t *testing.T) {
	status := "completed"
	repo := &PayoutRepo{
		ListFn: func(partnerID string, params structs.PayoutListParams) ([]database.Payout, int, error) {
			assert.Equal(t, &status, params.Status)
			return []database.Payout{}, 0, nil
		},
	}

	handler := queries.ListPayoutsHandler{Payouts: repo}
	_, err := handler.Handle("partner-1", structs.PayoutListParams{Status: &status})

	assert.NoError(t, err)
}

// ── AdminListPayouts ──

func TestAdminListPayouts_Success(t *testing.T) {
	payouts := []database.Payout{
		{Base: database.Base{ID: "p-1"}, Amount: 500, PartnerID: "partner-1"},
		{Base: database.Base{ID: "p-2"}, Amount: 300, PartnerID: "partner-2"},
		{Base: database.Base{ID: "p-3"}, Amount: 150, PartnerID: "partner-1"},
	}

	repo := &PayoutRepo{
		AdminListFn: func(params structs.PayoutListParams) ([]database.Payout, int, error) {
			return payouts, 3, nil
		},
	}

	handler := queries.AdminListPayoutsHandler{Payouts: repo}
	result, err := handler.Handle(structs.PayoutListParams{})

	assert.NoError(t, err)
	assert.Equal(t, 3, result.Total)
	assert.Len(t, result.Items, 3)
}

func TestAdminListPayouts_RepoError(t *testing.T) {
	repo := &PayoutRepo{
		AdminListFn: func(params structs.PayoutListParams) ([]database.Payout, int, error) {
			return nil, 0, errDB
		},
	}

	handler := queries.AdminListPayoutsHandler{Payouts: repo}
	result, err := handler.Handle(structs.PayoutListParams{})

	assert.ErrorIs(t, err, errDB)
	assert.Equal(t, 0, result.Total)
}

func TestAdminListPayouts_FilterByPartner(t *testing.T) {
	partnerID := "partner-1"
	repo := &PayoutRepo{
		AdminListFn: func(params structs.PayoutListParams) ([]database.Payout, int, error) {
			assert.Equal(t, &partnerID, params.PartnerID)
			return []database.Payout{}, 0, nil
		},
	}

	handler := queries.AdminListPayoutsHandler{Payouts: repo}
	_, err := handler.Handle(structs.PayoutListParams{PartnerID: &partnerID})

	assert.NoError(t, err)
}

// ── PayoutDetail ──

func TestPayoutDetail_Success(t *testing.T) {
	now := time.Now()
	payout := &database.Payout{
		Base:      database.Base{ID: "p-1"},
		PartnerID: "partner-1",
		Amount:    50.5,
		Currency:  "USDT",
		Status:    database.PayoutStatusPending,
		PeriodStart: now,
		PeriodEnd:   now,
	}
	items := []database.PayoutItem{
		{Base: database.Base{ID: "item-1"}, PayoutID: "p-1", CommissionID: "c-1", Amount: 20.0},
		{Base: database.Base{ID: "item-2"}, PayoutID: "p-1", CommissionID: "c-2", Amount: 30.5},
	}

	repo := &PayoutRepo{
		DetailFn: func(partnerID, id string) (*database.Payout, []database.PayoutItem, error) {
			assert.Equal(t, "partner-1", partnerID)
			assert.Equal(t, "p-1", id)
			return payout, items, nil
		},
	}

	handler := queries.PayoutDetailHandler{Payouts: repo}
	result, err := handler.Handle("partner-1", "p-1")

	assert.NoError(t, err)
	assert.Equal(t, "p-1", result.Payout.ID)
	assert.Equal(t, 50.5, result.Payout.Amount)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 20.0, result.Items[0].Amount)
	assert.Equal(t, 30.5, result.Items[1].Amount)
}

func TestPayoutDetail_NotFound(t *testing.T) {
	repo := &PayoutRepo{
		DetailFn: func(partnerID, id string) (*database.Payout, []database.PayoutItem, error) {
			return nil, nil, errDB
		},
	}

	handler := queries.PayoutDetailHandler{Payouts: repo}
	result, err := handler.Handle("partner-1", "missing")

	assert.ErrorIs(t, err, errDB)
	assert.Nil(t, result)
}

func TestPayoutDetail_NoItems(t *testing.T) {
	now := time.Now()
	payout := &database.Payout{
		Base:        database.Base{ID: "p-1"},
		PartnerID:   "partner-1",
		Amount:      0,
		Status:      database.PayoutStatusPending,
		PeriodStart: now,
		PeriodEnd:   now,
	}

	repo := &PayoutRepo{
		DetailFn: func(partnerID, id string) (*database.Payout, []database.PayoutItem, error) {
			return payout, []database.PayoutItem{}, nil
		},
	}

	handler := queries.PayoutDetailHandler{Payouts: repo}
	result, err := handler.Handle("partner-1", "p-1")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Items)
}

// ── AdminPayoutDetail ──

func TestAdminPayoutDetail_Success(t *testing.T) {
	now := time.Now()
	adminID := "admin-1"
	payout := &database.Payout{
		Base:        database.Base{ID: "p-1"},
		PartnerID:   "partner-1",
		Amount:      100,
		Status:      database.PayoutStatusProcessing,
		ApprovedBy:  &adminID,
		ProcessedAt: &now,
	}
	items := []database.PayoutItem{
		{Base: database.Base{ID: "item-1"}, PayoutID: "p-1", Amount: 100},
	}

	repo := &PayoutRepo{
		AdminDetailFn: func(id string) (*database.Payout, []database.PayoutItem, error) {
			assert.Equal(t, "p-1", id)
			return payout, items, nil
		},
	}

	handler := queries.AdminPayoutDetailHandler{Payouts: repo}
	result, err := handler.Handle("p-1")

	assert.NoError(t, err)
	assert.Equal(t, "p-1", result.Payout.ID)
	assert.Equal(t, database.PayoutStatusProcessing, result.Payout.Status)
	assert.Equal(t, &adminID, result.Payout.ApprovedBy)
	assert.Len(t, result.Items, 1)
}

func TestAdminPayoutDetail_NotFound(t *testing.T) {
	repo := &PayoutRepo{
		AdminDetailFn: func(id string) (*database.Payout, []database.PayoutItem, error) {
			return nil, nil, errDB
		},
	}

	handler := queries.AdminPayoutDetailHandler{Payouts: repo}
	result, err := handler.Handle("missing")

	assert.ErrorIs(t, err, errDB)
	assert.Nil(t, result)
}

// ── DTO ──

func TestPendingInfo_MinPayoutAmount(t *testing.T) {
	info := queries.MinPayoutAmount
	assert.Equal(t, 10.0, info)
}

func TestPayoutListParams_Filters(t *testing.T) {
	status := "pending"
	partnerID := "partner-1"
	params := structs.PayoutListParams{
		Status:    &status,
		PartnerID: &partnerID,
	}

	assert.Equal(t, "pending", *params.Status)
	assert.Equal(t, "partner-1", *params.PartnerID)
}

// ── PayoutItemRepo interface ──

func TestPayoutItemRepo_CreateBatch_Success(t *testing.T) {
	created := false
	repo := &PayoutItemRepo{
		CreateBatchFn: func(items []database.PayoutItem) error {
			created = true
			assert.Len(t, items, 2)
			return nil
		},
	}

	items := []database.PayoutItem{
		{PayoutID: "p-1", CommissionID: "c-1", Amount: 10},
		{PayoutID: "p-1", CommissionID: "c-2", Amount: 20},
	}
	err := repo.CreateBatch(items)

	assert.NoError(t, err)
	assert.True(t, created)
}

func TestPayoutItemRepo_CreateBatch_Error(t *testing.T) {
	repo := &PayoutItemRepo{
		CreateBatchFn: func(items []database.PayoutItem) error {
			return errDB
		},
	}

	err := repo.CreateBatch([]database.PayoutItem{{PayoutID: "p-1"}})

	assert.ErrorIs(t, err, errDB)
}

// ── PayoutRepo interface ──

func TestPayoutRepo_FindPendingByID_Success(t *testing.T) {
	repo := &PayoutRepo{
		FindPendingByIDFn: func(id string) (*database.Payout, error) {
			return &database.Payout{
				Base:   database.Base{ID: id},
				Status: database.PayoutStatusPending,
			}, nil
		},
	}

	payout, err := repo.FindPendingByID("p-1")

	assert.NoError(t, err)
	assert.Equal(t, "p-1", payout.ID)
	assert.Equal(t, database.PayoutStatusPending, payout.Status)
}

func TestPayoutRepo_FindPendingByID_NotFound(t *testing.T) {
	repo := &PayoutRepo{
		FindPendingByIDFn: func(id string) (*database.Payout, error) {
			return nil, errDB
		},
	}

	payout, err := repo.FindPendingByID("missing")

	assert.ErrorIs(t, err, errDB)
	assert.Nil(t, payout)
}

func TestPayoutRepo_Save(t *testing.T) {
	saved := false
	repo := &PayoutRepo{
		SaveFn: func(payout *database.Payout) error {
			saved = true
			return nil
		},
	}

	err := repo.Save(&database.Payout{Base: database.Base{ID: "p-1"}})

	assert.NoError(t, err)
	assert.True(t, saved)
}

func TestPayoutRepo_Create(t *testing.T) {
	created := false
	repo := &PayoutRepo{
		CreateFn: func(payout *database.Payout) error {
			created = true
			return nil
		},
	}

	err := repo.Create(&database.Payout{PartnerID: "partner-1", Amount: 50})

	assert.NoError(t, err)
	assert.True(t, created)
}

func TestPayoutRepo_Reload(t *testing.T) {
	repo := &PayoutRepo{
		ReloadFn: func(payout *database.Payout) error {
			payout.Partner = &database.Partner{Base: database.Base{ID: "partner-1"}}
			return nil
		},
	}

	payout := &database.Payout{Base: database.Base{ID: "p-1"}, PartnerID: "partner-1"}
	err := repo.Reload(payout)

	assert.NoError(t, err)
	assert.NotNil(t, payout.Partner)
	assert.Equal(t, "partner-1", payout.Partner.ID)
}
