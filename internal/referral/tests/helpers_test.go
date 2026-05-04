package tests

import (
	"errors"
	"testing"

	"xmeta-partner/database"
	"xmeta-partner/structs"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var errDB = errors.New("db error")

type ReferralLinkRepo struct {
	FindByCodeFn       func(code string) (*database.ReferralLink, error)
	FindByCodeActiveFn func(code string) (*database.ReferralLink, error)
	ListFn             func(partnerID string, params structs.ReferralListParams) ([]database.ReferralLink, int, error)
	CreateFn           func(link *database.ReferralLink) error
	CodeTakenFn        func(code string) (bool, error)
	CountByPartnerFn   func(partnerID string) (int64, error)
}

func (m *ReferralLinkRepo) FindByCode(code string) (*database.ReferralLink, error) {
	return m.FindByCodeFn(code)
}
func (m *ReferralLinkRepo) FindByCodeActive(code string) (*database.ReferralLink, error) {
	return m.FindByCodeActiveFn(code)
}
func (m *ReferralLinkRepo) List(partnerID string, params structs.ReferralListParams) ([]database.ReferralLink, int, error) {
	return m.ListFn(partnerID, params)
}
func (m *ReferralLinkRepo) Create(link *database.ReferralLink) error { return m.CreateFn(link) }
func (m *ReferralLinkRepo) CodeTaken(code string) (bool, error)     { return m.CodeTakenFn(code) }
func (m *ReferralLinkRepo) CountByPartner(partnerID string) (int64, error) {
	return m.CountByPartnerFn(partnerID)
}

type ReferralRepo struct {
	FindByIDFn           func(id string) (*database.Referral, error)
	FindByIDAndPartnerFn func(id, partnerID string) (*database.Referral, error)
	FindActiveByUserIDFn func(userID string) (*database.Referral, error)
	ListFn               func(partnerID string, params structs.ReferralListParams) ([]database.Referral, int, error)
	AdminListFn          func(params structs.AdminReferralListParams) ([]database.Referral, int, error)
	SaveFn               func(referral *database.Referral) error
	CountHistoryFn       func(userID string) int64
}

func (m *ReferralRepo) FindByID(id string) (*database.Referral, error) { return m.FindByIDFn(id) }
func (m *ReferralRepo) FindByIDAndPartner(id, partnerID string) (*database.Referral, error) {
	return m.FindByIDAndPartnerFn(id, partnerID)
}
func (m *ReferralRepo) FindActiveByUserID(userID string) (*database.Referral, error) {
	return m.FindActiveByUserIDFn(userID)
}
func (m *ReferralRepo) List(partnerID string, params structs.ReferralListParams) ([]database.Referral, int, error) {
	return m.ListFn(partnerID, params)
}
func (m *ReferralRepo) AdminList(params structs.AdminReferralListParams) ([]database.Referral, int, error) {
	return m.AdminListFn(params)
}
func (m *ReferralRepo) Save(referral *database.Referral) error { return m.SaveFn(referral) }
func (m *ReferralRepo) CountHistory(userID string) int64       { return m.CountHistoryFn(userID) }

func newTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm open: %v", err)
	}

	return gormDB, mock
}
