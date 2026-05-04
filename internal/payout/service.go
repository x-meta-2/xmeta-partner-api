package payout

import (
	"xmeta-partner/internal/payout/adapters"
	"xmeta-partner/internal/payout/app/commands"
	"xmeta-partner/internal/payout/app/queries"

	"gorm.io/gorm"
)

type Service struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	ApprovePayout       *commands.ApprovePayoutHandler
	RejectPayout        *commands.RejectPayoutHandler
	ProcessDailyPayouts *commands.ProcessDailyPayoutsHandler
}

type Queries struct {
	ListPayouts        *queries.ListPayoutsHandler
	PayoutDetail       *queries.PayoutDetailHandler
	PendingCommissions *queries.PendingCommissionsHandler
	AdminListPayouts   *queries.AdminListPayoutsHandler
	AdminPayoutDetail  *queries.AdminPayoutDetailHandler
}

func NewService(db *gorm.DB) *Service {
	payoutRepo := &adapters.GormPayoutRepo{DB: db}

	return &Service{
		Commands: Commands{
			ApprovePayout: &commands.ApprovePayoutHandler{
				DB:      db,
				Payouts: payoutRepo,
			},
			RejectPayout: &commands.RejectPayoutHandler{
				DB:      db,
				Payouts: payoutRepo,
			},
			ProcessDailyPayouts: &commands.ProcessDailyPayoutsHandler{
				DB: db,
			},
		},
		Queries: Queries{
			ListPayouts:        &queries.ListPayoutsHandler{Payouts: payoutRepo},
			PayoutDetail:       &queries.PayoutDetailHandler{Payouts: payoutRepo},
			PendingCommissions: &queries.PendingCommissionsHandler{DB: db},
			AdminListPayouts:   &queries.AdminListPayoutsHandler{Payouts: payoutRepo},
			AdminPayoutDetail:  &queries.AdminPayoutDetailHandler{Payouts: payoutRepo},
		},
	}
}
