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
