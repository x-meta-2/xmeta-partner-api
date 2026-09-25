package tests

import (
	"regexp"
	"testing"
	"time"

	"xmeta-partner/internal/analytics/adapters"
	"xmeta-partner/structs"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardSummary_UsesLiveReferralCounts(t *testing.T) {
	db, mock := newTestDB(t)
	now := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT * FROM "partners" WHERE id = $1 AND "partners"."deleted_at" IS NULL ORDER BY "partners"."id" LIMIT $2`,
	)).
		WithArgs("partner-1", 1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "deleted_at", "user_id", "tier_id", "status", "referral_code", "total_referrals", "total_earnings",
		}).AddRow("partner-1", now, now, nil, "user-1", "tier-1", "active", "UQKEBGP", 4, 12.5))

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT COALESCE(SUM(rebate_amount), 0) FROM "commissions" WHERE (partner_id = $1 AND status = $2) AND "commissions"."deleted_at" IS NULL`,
	)).
		WithArgs("partner-1", "pending").
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(1.25))

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT COALESCE(SUM(rebate_amount), 0) FROM "commissions" WHERE (partner_id = $1 AND trade_date >= $2 AND trade_date < $3) AND "commissions"."deleted_at" IS NULL`,
	)).
		WithArgs("partner-1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(2.5))

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT count(*) FROM "referrals" WHERE (partner_id = $1 AND ended_at IS NULL) AND "referrals"."deleted_at" IS NULL`,
	)).
		WithArgs("partner-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT count(*) FROM "referrals" WHERE (partner_id = $1 AND status = $2 AND ended_at IS NULL) AND "referrals"."deleted_at" IS NULL`,
	)).
		WithArgs("partner-1", "active").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT COALESCE(SUM(volume_usd), 0) FROM "commissions" WHERE partner_id = $1 AND "commissions"."deleted_at" IS NULL`,
	)).
		WithArgs("partner-1").
		WillReturnRows(sqlmock.NewRows([]string{"coalesce"}).AddRow(100.0))

	repo := adapters.GormDashboardRepo{DB: db}
	result, err := repo.GetSummary("partner-1", structs.DashboardSummaryParams{})

	require.NoError(t, err)
	assert.Equal(t, 12.5, result.TotalEarnings)
	assert.Equal(t, 3, result.TotalReferrals)
	assert.Equal(t, int64(0), result.ActiveReferrals)
	assert.Equal(t, 0.0, result.ConversionRate)
	assert.Equal(t, 1.25, result.PendingCommission)
	assert.Equal(t, 2.5, result.MonthEarnings)
	assert.Equal(t, 100.0, result.TotalVolume)
	assert.NoError(t, mock.ExpectationsWereMet())
}
