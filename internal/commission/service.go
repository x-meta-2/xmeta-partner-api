package commission

import (
	"xmeta-partner/internal/commission/adapters"
	"xmeta-partner/internal/commission/app/commands"
	"xmeta-partner/internal/commission/app/queries"

	"gorm.io/gorm"
)

type Service struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	ProcessTradeEvent *commands.ProcessTradeEventHandler
}

type Queries struct {
	ListCommissions      *queries.ListCommissionsHandler
	AdminListCommissions *queries.AdminListCommissionsHandler
	CommissionBreakdown  *queries.CommissionBreakdownHandler
	DailySummary         *queries.DailySummaryHandler
}

func NewService(db *gorm.DB) *Service {
	commissionRepo := &adapters.GormCommissionRepo{DB: db}
	tradeEventRepo := &adapters.GormTradeEventRepo{DB: db}

	return &Service{
		Commands: Commands{
			ProcessTradeEvent: &commands.ProcessTradeEventHandler{Repo: tradeEventRepo},
		},
		Queries: Queries{
			ListCommissions:      &queries.ListCommissionsHandler{Commissions: commissionRepo},
			AdminListCommissions: &queries.AdminListCommissionsHandler{Commissions: commissionRepo},
			CommissionBreakdown:  &queries.CommissionBreakdownHandler{Commissions: commissionRepo},
			DailySummary:         &queries.DailySummaryHandler{Commissions: commissionRepo},
		},
	}
}
