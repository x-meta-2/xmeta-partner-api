package port

import (
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/commission/app/dto"
	"xmeta-partner/structs"
)

type CommissionRepo interface {
	List(partnerID string, params structs.CommissionListParams) ([]database.Commission, int, error)
	AdminList(params structs.AdminCommissionListParams) ([]database.Commission, int, error)
	Breakdown(partnerID string, params structs.CommissionBreakdownParams) (float64, error)
	DailySummary(partnerID string, params structs.ChartParams) ([]dto.DailyItem, error)
}

type TradeEventRepo interface {
	ExistsByPositionID(positionID string) (bool, error)
	IsUserKycVerified(userID string) (bool, error)
	FindActiveReferral(userID string, tradeDate time.Time) (*database.Referral, error)
	FindActivePartnerWithTier(partnerID string) (*database.Partner, error)
	CreateCommission(c *database.Commission) error
	IncrementPartnerEarnings(partnerID string, amount float64) error
	ActivateReferral(referralID string, firstTradeAt time.Time) error
	GetPartnerTotalVolume(partnerID string) (float64, error)
	GetPartnerActiveClients(partnerID string) (int64, error)
	FindAllTiersAsc() ([]database.PartnerTier, error)
	UpgradePartnerTier(partnerID string, newTierID string, newLevel int) error
	RunInTx(fn func(TradeEventRepo) error) error
}
