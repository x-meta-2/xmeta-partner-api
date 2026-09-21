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

type PayoutRepo struct {
	FindPendingByIDFn func(id string) (*database.Payout, error)
	ListFn            func(partnerID string, params structs.PayoutListParams) ([]database.Payout, int, error)
	AdminListFn       func(params structs.PayoutListParams) ([]database.Payout, int, error)
	DetailFn          func(partnerID, id string) (*database.Payout, []database.PayoutItem, error)
	AdminDetailFn     func(id string) (*database.Payout, []database.PayoutItem, error)
	SaveFn            func(payout *database.Payout) error
	CreateFn          func(payout *database.Payout) error
	ReloadFn          func(payout *database.Payout) error
}

func (m *PayoutRepo) FindPendingByID(id string) (*database.Payout, error) {
	return m.FindPendingByIDFn(id)
}
func (m *PayoutRepo) List(partnerID string, params structs.PayoutListParams) ([]database.Payout, int, error) {
	return m.ListFn(partnerID, params)
}
func (m *PayoutRepo) AdminList(params structs.PayoutListParams) ([]database.Payout, int, error) {
	return m.AdminListFn(params)
}
func (m *PayoutRepo) Detail(partnerID, id string) (*database.Payout, []database.PayoutItem, error) {
	return m.DetailFn(partnerID, id)
}
func (m *PayoutRepo) AdminDetail(id string) (*database.Payout, []database.PayoutItem, error) {
	return m.AdminDetailFn(id)
}
func (m *PayoutRepo) Save(payout *database.Payout) error   { return m.SaveFn(payout) }
func (m *PayoutRepo) Create(payout *database.Payout) error { return m.CreateFn(payout) }
func (m *PayoutRepo) Reload(payout *database.Payout) error { return m.ReloadFn(payout) }

type PayoutItemRepo struct {
	CreateBatchFn func(items []database.PayoutItem) error
}

func (m *PayoutItemRepo) CreateBatch(items []database.PayoutItem) error {
	return m.CreateBatchFn(items)
}
