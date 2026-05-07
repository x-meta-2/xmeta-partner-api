package tests

import (
	"regexp"
	"testing"

	"xmeta-partner/database"
	"xmeta-partner/internal/referral/app/commands"
	"xmeta-partner/internal/referral/domain"
	"xmeta-partner/structs"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// CreateLink
// ---------------------------------------------------------------------------

func TestCreateLink_WithCustomCode(t *testing.T) {
	var created *database.ReferralLink

	links := &ReferralLinkRepo{
		CountByPartnerFn: func(partnerID string) (int64, error) { return 0, nil },
		CodeTakenFn:      func(code string) (bool, error) { return false, nil },
		CreateFn: func(link *database.ReferralLink) error {
			created = link
			return nil
		},
	}

	handler := commands.CreateLinkHandler{Links: links}
	result, err := handler.Handle("partner-1", structs.ReferralLinkCreateParams{
		Code: "MYCODE",
	})

	assert.NoError(t, err)
	assert.Equal(t, "MYCODE", result.Code)
	assert.Equal(t, "partner-1", result.PartnerID)
	assert.True(t, result.IsActive)
	assert.Contains(t, result.URL, "MYCODE")
	assert.Equal(t, created.Code, result.Code)
}

func TestCreateLink_MaxLinksReached(t *testing.T) {
	links := &ReferralLinkRepo{
		CountByPartnerFn: func(partnerID string) (int64, error) {
			return int64(domain.MaxReferralLinksPerPartner), nil
		},
	}

	handler := commands.CreateLinkHandler{Links: links}
	_, err := handler.Handle("partner-1", structs.ReferralLinkCreateParams{Code: "ABC"})

	assert.ErrorIs(t, err, domain.ErrMaxLinksReached)
}

func TestCreateLink_CodeAlreadyTaken(t *testing.T) {
	links := &ReferralLinkRepo{
		CountByPartnerFn: func(partnerID string) (int64, error) { return 0, nil },
		CodeTakenFn:      func(code string) (bool, error) { return true, nil },
	}

	handler := commands.CreateLinkHandler{Links: links}
	_, err := handler.Handle("partner-1", structs.ReferralLinkCreateParams{Code: "TAKEN1"})

	assert.ErrorIs(t, err, domain.ErrCodeTaken)
}

func TestCreateLink_AutoGenerateCode(t *testing.T) {
	links := &ReferralLinkRepo{
		CountByPartnerFn: func(partnerID string) (int64, error) { return 0, nil },
		CodeTakenFn:      func(code string) (bool, error) { return false, nil },
		CreateFn:         func(link *database.ReferralLink) error { return nil },
	}

	handler := commands.CreateLinkHandler{Links: links}
	result, err := handler.Handle("partner-1", structs.ReferralLinkCreateParams{})

	assert.NoError(t, err)
	assert.NotEmpty(t, result.Code)
	assert.True(t, result.IsActive)
}

func TestCreateLink_AllAutoCodesCollide(t *testing.T) {
	links := &ReferralLinkRepo{
		CountByPartnerFn: func(partnerID string) (int64, error) { return 0, nil },
		CodeTakenFn:      func(code string) (bool, error) { return true, nil },
	}

	handler := commands.CreateLinkHandler{Links: links}
	_, err := handler.Handle("partner-1", structs.ReferralLinkCreateParams{})

	assert.ErrorIs(t, err, domain.ErrCodeGenerationFailed)
}

func TestCreateLink_CustomURL(t *testing.T) {
	links := &ReferralLinkRepo{
		CountByPartnerFn: func(partnerID string) (int64, error) { return 0, nil },
		CodeTakenFn:      func(code string) (bool, error) { return false, nil },
		CreateFn:         func(link *database.ReferralLink) error { return nil },
	}

	handler := commands.CreateLinkHandler{Links: links}
	result, err := handler.Handle("partner-1", structs.ReferralLinkCreateParams{
		Code: "CUSTOM",
		URL:  "https://example.com/promo",
	})

	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/promo", result.URL)
}

func TestCreateLink_CreateFails(t *testing.T) {
	links := &ReferralLinkRepo{
		CountByPartnerFn: func(partnerID string) (int64, error) { return 0, nil },
		CodeTakenFn:      func(code string) (bool, error) { return false, nil },
		CreateFn:         func(link *database.ReferralLink) error { return errDB },
	}

	handler := commands.CreateLinkHandler{Links: links}
	_, err := handler.Handle("partner-1", structs.ReferralLinkCreateParams{Code: "VALID1"})

	assert.ErrorIs(t, err, errDB)
}

// ---------------------------------------------------------------------------
// LinkReferral
// ---------------------------------------------------------------------------

func TestLinkReferral_LinkNotFound(t *testing.T) {
	gormDB, _ := newTestDB(t)

	links := &ReferralLinkRepo{
		FindByCodeActiveFn: func(code string) (*database.ReferralLink, error) {
			return nil, domain.ErrLinkNotFound
		},
	}

	handler := commands.LinkReferralHandler{DB: gormDB, Links: links}
	err := handler.Handle("user-1", "BADCODE")

	assert.ErrorIs(t, err, domain.ErrLinkNotFound)
}

func TestLinkReferral_PartnerNotActive(t *testing.T) {
	gormDB, mock := newTestDB(t)

	links := &ReferralLinkRepo{
		FindByCodeActiveFn: func(code string) (*database.ReferralLink, error) {
			return &database.ReferralLink{
				Base:      database.Base{ID: "link-1"},
				PartnerID: "partner-1",
				Code:      code,
				IsActive:  true,
			}, nil
		},
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "status"}).
		AddRow("partner-1", "owner-1", "suspended")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "partners" WHERE id = $1 AND "partners"."deleted_at" IS NULL ORDER BY "partners"."id" LIMIT $2`)).
		WithArgs("partner-1", 1).
		WillReturnRows(rows)

	handler := commands.LinkReferralHandler{DB: gormDB, Links: links}
	err := handler.Handle("user-2", "TESTCODE")

	assert.ErrorIs(t, err, domain.ErrPartnerNotActive)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkReferral_SelfReferral(t *testing.T) {
	gormDB, mock := newTestDB(t)

	links := &ReferralLinkRepo{
		FindByCodeActiveFn: func(code string) (*database.ReferralLink, error) {
			return &database.ReferralLink{
				Base:      database.Base{ID: "link-1"},
				PartnerID: "partner-1",
				Code:      code,
				IsActive:  true,
			}, nil
		},
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "status"}).
		AddRow("partner-1", "user-self", "active")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "partners" WHERE id = $1 AND "partners"."deleted_at" IS NULL ORDER BY "partners"."id" LIMIT $2`)).
		WithArgs("partner-1", 1).
		WillReturnRows(rows)

	handler := commands.LinkReferralHandler{DB: gormDB, Links: links}
	err := handler.Handle("user-self", "TESTCODE")

	assert.ErrorIs(t, err, domain.ErrSelfReferral)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkReferral_SamePartnerRetry_Idempotent(t *testing.T) {
	gormDB, mock := newTestDB(t)

	links := &ReferralLinkRepo{
		FindByCodeActiveFn: func(code string) (*database.ReferralLink, error) {
			return &database.ReferralLink{
				Base:      database.Base{ID: "link-1"},
				PartnerID: "partner-1",
				Code:      code,
				IsActive:  true,
			}, nil
		},
	}

	partnerRows := sqlmock.NewRows([]string{"id", "user_id", "status"}).
		AddRow("partner-1", "owner-1", "active")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "partners" WHERE id = $1 AND "partners"."deleted_at" IS NULL ORDER BY "partners"."id" LIMIT $2`)).
		WithArgs("partner-1", 1).
		WillReturnRows(partnerRows)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`SELECT pg_advisory_xact_lock`)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	existingRows := sqlmock.NewRows([]string{"id", "partner_id", "referred_user_id", "referral_link_id", "status"}).
		AddRow("ref-existing", "partner-1", "user-2", "link-1", "registered")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "referrals" WHERE (referred_user_id = $1 AND partner_id = $2 AND ended_at IS NULL) AND "referrals"."deleted_at" IS NULL ORDER BY "referrals"."id" LIMIT $3`)).
		WithArgs("user-2", "partner-1", 1).
		WillReturnRows(existingRows)

	mock.ExpectCommit()

	handler := commands.LinkReferralHandler{DB: gormDB, Links: links}
	err := handler.Handle("user-2", "TESTCODE")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkReferral_SamePartnerDifferentLink_UpdatesLinkID(t *testing.T) {
	gormDB, mock := newTestDB(t)

	links := &ReferralLinkRepo{
		FindByCodeActiveFn: func(code string) (*database.ReferralLink, error) {
			return &database.ReferralLink{
				Base:      database.Base{ID: "link-new"},
				PartnerID: "partner-1",
				Code:      code,
				IsActive:  true,
			}, nil
		},
	}

	partnerRows := sqlmock.NewRows([]string{"id", "user_id", "status"}).
		AddRow("partner-1", "owner-1", "active")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "partners" WHERE id = $1 AND "partners"."deleted_at" IS NULL ORDER BY "partners"."id" LIMIT $2`)).
		WithArgs("partner-1", 1).
		WillReturnRows(partnerRows)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`SELECT pg_advisory_xact_lock`)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	existingRows := sqlmock.NewRows([]string{"id", "partner_id", "referred_user_id", "referral_link_id", "status"}).
		AddRow("ref-existing", "partner-1", "user-2", "link-old", "registered")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "referrals" WHERE (referred_user_id = $1 AND partner_id = $2 AND ended_at IS NULL) AND "referrals"."deleted_at" IS NULL ORDER BY "referrals"."id" LIMIT $3`)).
		WithArgs("user-2", "partner-1", 1).
		WillReturnRows(existingRows)

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "referrals" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	handler := commands.LinkReferralHandler{DB: gormDB, Links: links}
	err := handler.Handle("user-2", "NEWCODE")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkReferral_IdempotencyCheckDBError_Returns(t *testing.T) {
	gormDB, mock := newTestDB(t)

	links := &ReferralLinkRepo{
		FindByCodeActiveFn: func(code string) (*database.ReferralLink, error) {
			return &database.ReferralLink{
				Base:      database.Base{ID: "link-1"},
				PartnerID: "partner-1",
				Code:      code,
				IsActive:  true,
			}, nil
		},
	}

	partnerRows := sqlmock.NewRows([]string{"id", "user_id", "status"}).
		AddRow("partner-1", "owner-1", "active")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "partners" WHERE id = $1 AND "partners"."deleted_at" IS NULL ORDER BY "partners"."id" LIMIT $2`)).
		WithArgs("partner-1", 1).
		WillReturnRows(partnerRows)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`SELECT pg_advisory_xact_lock`)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "referrals" WHERE (referred_user_id = $1 AND partner_id = $2 AND ended_at IS NULL) AND "referrals"."deleted_at" IS NULL ORDER BY "referrals"."id" LIMIT $3`)).
		WithArgs("user-2", "partner-1", 1).
		WillReturnError(errDB)

	mock.ExpectRollback()

	handler := commands.LinkReferralHandler{DB: gormDB, Links: links}
	err := handler.Handle("user-2", "TESTCODE")

	assert.ErrorIs(t, err, errDB)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLinkReferral_DifferentPartner_ClosesOld(t *testing.T) {
	gormDB, mock := newTestDB(t)

	links := &ReferralLinkRepo{
		FindByCodeActiveFn: func(code string) (*database.ReferralLink, error) {
			return &database.ReferralLink{
				Base:      database.Base{ID: "link-2"},
				PartnerID: "partner-2",
				Code:      code,
				IsActive:  true,
			}, nil
		},
	}

	partnerRows := sqlmock.NewRows([]string{"id", "user_id", "status"}).
		AddRow("partner-2", "owner-2", "active")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "partners" WHERE id = $1 AND "partners"."deleted_at" IS NULL ORDER BY "partners"."id" LIMIT $2`)).
		WithArgs("partner-2", 1).
		WillReturnRows(partnerRows)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`SELECT pg_advisory_xact_lock`)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "referrals" WHERE (referred_user_id = $1 AND partner_id = $2 AND ended_at IS NULL) AND "referrals"."deleted_at" IS NULL ORDER BY "referrals"."id" LIMIT $3`)).
		WithArgs("user-2", "partner-2", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "referrals" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "referrals"`)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "referrals"`)).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectCommit()

	handler := commands.LinkReferralHandler{DB: gormDB, Links: links}
	err := handler.Handle("user-2", "NEWCODE")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// ---------------------------------------------------------------------------
// UnlinkReferral
// ---------------------------------------------------------------------------

func TestUnlinkReferral_Success(t *testing.T) {
	gormDB, mock := newTestDB(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "referrals" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	handler := commands.UnlinkReferralHandler{DB: gormDB}
	err := handler.Handle("user-1")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUnlinkReferral_NoActive(t *testing.T) {
	gormDB, mock := newTestDB(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "referrals" SET`)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	handler := commands.UnlinkReferralHandler{DB: gormDB}
	err := handler.Handle("user-1")

	assert.ErrorIs(t, err, domain.ErrNoActiveReferral)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUnlinkReferral_DBError(t *testing.T) {
	gormDB, mock := newTestDB(t)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "referrals" SET`)).
		WillReturnError(errDB)
	mock.ExpectRollback()

	handler := commands.UnlinkReferralHandler{DB: gormDB}
	err := handler.Handle("user-1")

	assert.ErrorIs(t, err, errDB)
	assert.NoError(t, mock.ExpectationsWereMet())
}
