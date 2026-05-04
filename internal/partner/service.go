package partner

import (
	"xmeta-partner/internal/partner/adapters"
	"xmeta-partner/internal/partner/app/commands"
	"xmeta-partner/internal/partner/app/queries"

	"gorm.io/gorm"
)

type Service struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	ApproveApplication  *commands.ApproveApplicationHandler
	RejectApplication   *commands.RejectApplicationHandler
	UpdatePartnerTier   *commands.UpdatePartnerTierHandler
	UpdatePartnerStatus *commands.UpdatePartnerStatusHandler
	CreateTier          *commands.CreateTierHandler
	UpdateTier          *commands.UpdateTierHandler
	DeleteTier          *commands.DeleteTierHandler
}

type Queries struct {
	ListPartners     *queries.ListPartnersHandler
	GetPartnerDetail *queries.GetPartnerDetailHandler
	ListApplications *queries.ListApplicationsHandler
	GetApplication   *queries.GetApplicationHandler
	ListTiers        *queries.ListTiersHandler
}

func NewService(db *gorm.DB) *Service {
	partnerRepo := &adapters.GormPartnerRepo{DB: db}
	appRepo := &adapters.GormApplicationRepo{DB: db}
	tierRepo := &adapters.GormTierRepo{DB: db}

	return &Service{
		Commands: Commands{
			ApproveApplication: &commands.ApproveApplicationHandler{
				DB:       db,
				Apps:     appRepo,
				Tiers:    tierRepo,
				Partners: partnerRepo,
			},
			RejectApplication: &commands.RejectApplicationHandler{
				Apps: appRepo,
			},
			UpdatePartnerTier: &commands.UpdatePartnerTierHandler{
				Partners: partnerRepo,
				Tiers:    tierRepo,
			},
			UpdatePartnerStatus: &commands.UpdatePartnerStatusHandler{
				Partners: partnerRepo,
			},
			CreateTier: &commands.CreateTierHandler{
				Tiers: tierRepo,
			},
			UpdateTier: &commands.UpdateTierHandler{
				Tiers: tierRepo,
			},
			DeleteTier: &commands.DeleteTierHandler{
				Tiers: tierRepo,
			},
		},
		Queries: Queries{
			ListPartners:     &queries.ListPartnersHandler{Partners: partnerRepo},
			GetPartnerDetail: &queries.GetPartnerDetailHandler{Detail: partnerRepo},
			ListApplications: &queries.ListApplicationsHandler{Apps: appRepo},
			GetApplication:   &queries.GetApplicationHandler{Apps: appRepo},
			ListTiers:        &queries.ListTiersHandler{Tiers: tierRepo},
		},
	}
}
