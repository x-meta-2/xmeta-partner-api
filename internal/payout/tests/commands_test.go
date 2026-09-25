package tests

import (
	"regexp"
	"testing"

	"xmeta-partner/database"
	"xmeta-partner/internal/payout/app/commands"
	"xmeta-partner/internal/payout/domain"
	"xmeta-partner/structs"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// ── Domain constants ──

func TestPayoutStatus_Constants(t *testing.T) {
	assert.Equal(t, domain.PayoutStatus("pending"), domain.StatusPending)
	assert.Equal(t, domain.PayoutStatus("processing"), domain.StatusProcessing)
	assert.Equal(t, domain.PayoutStatus("completed"), domain.StatusCompleted)
	assert.Equal(t, domain.PayoutStatus("failed"), domain.StatusFailed)
}

func TestPayoutStatus_MatchesDatabase(t *testing.T) {
	assert.Equal(t, string(domain.StatusPending), string(database.PayoutStatusPending))
	assert.Equal(t, string(domain.StatusProcessing), string(database.PayoutStatusProcessing))
	assert.Equal(t, string(domain.StatusCompleted), string(database.PayoutStatusCompleted))
	assert.Equal(t, string(domain.StatusFailed), string(database.PayoutStatusFailed))
}

func TestErrPayoutNotFound(t *testing.T) {
	assert.EqualError(t, domain.ErrPayoutNotFound, "payout not found or already processed")
}

func TestPayoutReviewParams(t *testing.T) {
	params := structs.PayoutReviewParams{FailureReason: "insufficient documents"}
	assert.Equal(t, "insufficient documents", params.FailureReason)
}

func TestApprovePayout_MovesPendingToProcessing(t *testing.T) {
	gormDB, mock := newTestDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payouts"`)).
		WithArgs("payout-1", "pending", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).
			AddRow("payout-1", string(database.PayoutStatusPending)))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "payouts" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	repo := &PayoutRepo{
		ReloadFn: func(payout *database.Payout) error {
			assert.Equal(t, database.PayoutStatusProcessing, payout.Status)
			return nil
		},
	}
	handler := commands.ApprovePayoutHandler{DB: gormDB, Payouts: repo}

	payout, err := handler.Handle("payout-1", "admin-1")

	assert.NoError(t, err)
	assert.Equal(t, database.PayoutStatusProcessing, payout.Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompletePayout_MarksPayoutAndCommissionsPaid(t *testing.T) {
	gormDB, mock := newTestDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payouts"`)).
		WithArgs("payout-1", database.PayoutStatusProcessing, 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "status"}).
			AddRow("payout-1", string(database.PayoutStatusProcessing)))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "payouts" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "commissions" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	repo := &PayoutRepo{
		ReloadFn: func(payout *database.Payout) error {
			assert.Equal(t, database.PayoutStatusCompleted, payout.Status)
			assert.Equal(t, "tx-123", payout.TransactionID)
			return nil
		},
	}
	handler := commands.CompletePayoutHandler{DB: gormDB, Payouts: repo}

	payout, err := handler.Handle("payout-1", structs.PayoutCompleteParams{TransactionID: " tx-123 "})

	assert.NoError(t, err)
	assert.Equal(t, database.PayoutStatusCompleted, payout.Status)
	assert.Equal(t, "tx-123", payout.TransactionID)
	assert.NoError(t, mock.ExpectationsWereMet())
}
