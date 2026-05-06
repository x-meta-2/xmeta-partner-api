package tests

import (
	"errors"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/commission/app/dto"
	"xmeta-partner/structs"
)

var errDB = errors.New("db error")

// ─── CommissionRepo mock ───

type CommissionRepo struct {
	ListFn         func(partnerID string, params structs.CommissionListParams) ([]database.Commission, int, error)
	AdminListFn    func(params structs.AdminCommissionListParams) ([]database.Commission, int, error)
	BreakdownFn    func(partnerID string, params structs.CommissionBreakdownParams) (float64, error)
	DailySummaryFn func(partnerID string, params structs.ChartParams) ([]dto.DailyItem, error)
}

func (m *CommissionRepo) List(partnerID string, params structs.CommissionListParams) ([]database.Commission, int, error) {
	return m.ListFn(partnerID, params)
}
func (m *CommissionRepo) AdminList(params structs.AdminCommissionListParams) ([]database.Commission, int, error) {
	return m.AdminListFn(params)
}
func (m *CommissionRepo) Breakdown(partnerID string, params structs.CommissionBreakdownParams) (float64, error) {
	return m.BreakdownFn(partnerID, params)
}
func (m *CommissionRepo) DailySummary(partnerID string, params structs.ChartParams) ([]dto.DailyItem, error) {
	return m.DailySummaryFn(partnerID, params)
}

// ─── TradeEventRepo mock ───

type TradeEventRepo struct {
	ExistsByPositionIDFn        func(positionID string) (bool, error)
	IsUserKycVerifiedFn         func(userID string) (bool, error)
	FindActiveReferralFn        func(userID string, tradeDate time.Time) (*database.Referral, error)
	FindActivePartnerWithTierFn func(partnerID string) (*database.Partner, error)
	CreateCommissionFn          func(c *database.Commission) error
	IncrementPartnerEarningsFn  func(partnerID string, amount float64) error
	ActivateReferralFn          func(referralID string, firstTradeAt time.Time) error
}

func (m *TradeEventRepo) ExistsByPositionID(positionID string) (bool, error) {
	return m.ExistsByPositionIDFn(positionID)
}
func (m *TradeEventRepo) IsUserKycVerified(userID string) (bool, error) {
	return m.IsUserKycVerifiedFn(userID)
}
func (m *TradeEventRepo) FindActiveReferral(userID string, tradeDate time.Time) (*database.Referral, error) {
	return m.FindActiveReferralFn(userID, tradeDate)
}
func (m *TradeEventRepo) FindActivePartnerWithTier(partnerID string) (*database.Partner, error) {
	return m.FindActivePartnerWithTierFn(partnerID)
}
func (m *TradeEventRepo) CreateCommission(c *database.Commission) error {
	return m.CreateCommissionFn(c)
}
func (m *TradeEventRepo) IncrementPartnerEarnings(partnerID string, amount float64) error {
	return m.IncrementPartnerEarningsFn(partnerID, amount)
}
func (m *TradeEventRepo) ActivateReferral(referralID string, firstTradeAt time.Time) error {
	return m.ActivateReferralFn(referralID, firstTradeAt)
}
