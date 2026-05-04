package tests

import (
	"errors"

	"xmeta-partner/database"
	"xmeta-partner/internal/partner/app/dto"
	"xmeta-partner/structs"
)

var errDB = errors.New("db error")

type ApplicationRepo struct {
	FindByIDFn        func(id string) (*database.PartnerApplication, error)
	FindPendingByIDFn func(id string) (*database.PartnerApplication, error)
	ListFn            func(params structs.ApplicationListParams) ([]database.PartnerApplication, int, error)
	SaveFn            func(app *database.PartnerApplication) error
}

func (m *ApplicationRepo) FindByID(id string) (*database.PartnerApplication, error) {
	return m.FindByIDFn(id)
}
func (m *ApplicationRepo) FindPendingByID(id string) (*database.PartnerApplication, error) {
	return m.FindPendingByIDFn(id)
}
func (m *ApplicationRepo) List(params structs.ApplicationListParams) ([]database.PartnerApplication, int, error) {
	return m.ListFn(params)
}
func (m *ApplicationRepo) Save(app *database.PartnerApplication) error {
	return m.SaveFn(app)
}

type TierRepo struct {
	FindByIDFn           func(id string) (*database.PartnerTier, error)
	FindDefaultFn        func() (*database.PartnerTier, error)
	ListFn               func() ([]database.PartnerTier, error)
	CreateFn             func(tier *database.PartnerTier) error
	UpdateFn             func(id string, fields map[string]interface{}) error
	DeleteFn             func(id string) error
	CountPartnersByTierFn func(tierID string) (int64, error)
	ClearDefaultExceptFn func(tierID string) error
}

func (m *TierRepo) FindByID(id string) (*database.PartnerTier, error) {
	return m.FindByIDFn(id)
}
func (m *TierRepo) FindDefault() (*database.PartnerTier, error) { return m.FindDefaultFn() }
func (m *TierRepo) List() ([]database.PartnerTier, error)       { return m.ListFn() }
func (m *TierRepo) Create(tier *database.PartnerTier) error     { return m.CreateFn(tier) }
func (m *TierRepo) Update(id string, fields map[string]interface{}) error {
	return m.UpdateFn(id, fields)
}
func (m *TierRepo) Delete(id string) error                      { return m.DeleteFn(id) }
func (m *TierRepo) CountPartnersByTier(tierID string) (int64, error) {
	return m.CountPartnersByTierFn(tierID)
}
func (m *TierRepo) ClearDefaultExcept(tierID string) error { return m.ClearDefaultExceptFn(tierID) }

type PartnerRepo struct {
	FindByIDFn     func(id string) (*database.Partner, error)
	ListFn         func(params structs.PartnerListParams) ([]database.Partner, int, error)
	UpdateTierFn   func(id string, tierID string) error
	UpdateStatusFn func(id string, status string) error
}

func (m *PartnerRepo) FindByID(id string) (*database.Partner, error) { return m.FindByIDFn(id) }
func (m *PartnerRepo) List(params structs.PartnerListParams) ([]database.Partner, int, error) {
	return m.ListFn(params)
}
func (m *PartnerRepo) UpdateTier(id string, tierID string) error { return m.UpdateTierFn(id, tierID) }
func (m *PartnerRepo) UpdateStatus(id string, status string) error {
	return m.UpdateStatusFn(id, status)
}

type PartnerDetailRepo struct {
	GetDetailFn func(id string) (*dto.PartnerDetail, error)
}

func (m *PartnerDetailRepo) GetDetail(id string) (*dto.PartnerDetail, error) {
	return m.GetDetailFn(id)
}
