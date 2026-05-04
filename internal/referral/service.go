package referral

import (
	"xmeta-partner/internal/referral/adapters"
	"xmeta-partner/internal/referral/app/commands"
	"xmeta-partner/internal/referral/app/queries"

	"gorm.io/gorm"
)

type Service struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	LinkReferral   *commands.LinkReferralHandler
	UnlinkReferral *commands.UnlinkReferralHandler
	CreateLink     *commands.CreateLinkHandler
}

type Queries struct {
	ListReferrals       *queries.ListReferralsHandler
	ReferralDetail      *queries.ReferralDetailHandler
	ReferralStats       *queries.ReferralStatsHandler
	ListLinks           *queries.ListLinksHandler
	LookupLink          *queries.LookupLinkHandler
	AdminListReferrals  *queries.AdminListReferralsHandler
	AdminReferralDetail *queries.AdminReferralDetailHandler
}

func NewService(db *gorm.DB) *Service {
	referralRepo := &adapters.GormReferralRepo{DB: db}
	linkRepo := &adapters.GormReferralLinkRepo{DB: db}

	return &Service{
		Commands: Commands{
			LinkReferral: &commands.LinkReferralHandler{
				DB:    db,
				Links: linkRepo,
			},
			UnlinkReferral: &commands.UnlinkReferralHandler{
				DB: db,
			},
			CreateLink: &commands.CreateLinkHandler{
				Links: linkRepo,
			},
		},
		Queries: Queries{
			ListReferrals:  &queries.ListReferralsHandler{Referrals: referralRepo},
			ReferralDetail: &queries.ReferralDetailHandler{Referrals: referralRepo},
			ReferralStats:  &queries.ReferralStatsHandler{DB: db},
			ListLinks:      &queries.ListLinksHandler{Links: linkRepo},
			LookupLink:     &queries.LookupLinkHandler{DB: db},
			AdminListReferrals: &queries.AdminListReferralsHandler{
				Referrals: referralRepo,
			},
			AdminReferralDetail: &queries.AdminReferralDetailHandler{
				DB:        db,
				Referrals: referralRepo,
			},
		},
	}
}
