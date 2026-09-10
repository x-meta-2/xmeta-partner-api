package tests

import (
	"regexp"
	"testing"

	"xmeta-partner/internal/referral/app/queries"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/gorm"
)

func TestCheckDirectReferral(t *testing.T) {
	t.Run("returns true for the active owner's direct referral", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "partners" WHERE (user_id = $1 AND status = $2) AND "partners"."deleted_at" IS NULL ORDER BY "partners"."id" LIMIT $3`)).
			WithArgs("owner-uid", "active", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "status"}).AddRow("partner-id", "owner-uid", "active"))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "referrals" WHERE (partner_id = $1 AND referred_user_id = $2 AND ended_at IS NULL) AND "referrals"."deleted_at" IS NULL`)).
			WithArgs("partner-id", "target-uid").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		result, err := (&queries.CheckDirectReferralHandler{DB: db}).Handle("owner-uid", "target-uid")
		if err != nil || !result {
			t.Fatalf("result=%v err=%v", result, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("returns false for a non-partner owner", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "partners" WHERE (user_id = $1 AND status = $2) AND "partners"."deleted_at" IS NULL ORDER BY "partners"."id" LIMIT $3`)).
			WithArgs("owner-uid", "active", 1).WillReturnError(gorm.ErrRecordNotFound)

		result, err := (&queries.CheckDirectReferralHandler{DB: db}).Handle("owner-uid", "target-uid")
		if err != nil || result {
			t.Fatalf("result=%v err=%v", result, err)
		}
	})
}
