package port

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/partner/app/dto"
	"xmeta-partner/structs"
)

type PartnerRepo interface {
	FindByID(id string) (*database.Partner, error)
	List(params structs.PartnerListParams) ([]database.Partner, int, error)
	UpdateTier(id string, tierID string) error
	UpdateStatus(id string, status string) error
}

type ApplicationRepo interface {
	FindByID(id string) (*database.PartnerApplication, error)
	FindPendingByID(id string) (*database.PartnerApplication, error)
	List(params structs.ApplicationListParams) ([]database.PartnerApplication, int, error)
	Save(app *database.PartnerApplication) error
}

type TierRepo interface {
	FindByID(id string) (*database.PartnerTier, error)
	FindDefault() (*database.PartnerTier, error)
	List() ([]database.PartnerTier, error)
	Create(tier *database.PartnerTier) error
	Update(id string, fields map[string]interface{}) error
	Delete(id string) error
	CountPartnersByTier(tierID string) (int64, error)
	ClearDefaultExcept(tierID string) error
	RunInTx(fn func(TierRepo) error) error
}

type PartnerDetailRepo interface {
	GetDetail(id string) (*dto.PartnerDetail, error)
}
