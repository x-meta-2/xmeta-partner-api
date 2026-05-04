package tests

import (
	"testing"

	"xmeta-partner/database"
	"xmeta-partner/internal/partner/app/commands"
	"xmeta-partner/internal/partner/domain"

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
