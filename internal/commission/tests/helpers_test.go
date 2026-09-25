package tests

import (
	"errors"

	"xmeta-partner/database"
	"xmeta-partner/internal/commission/app/dto"
	"xmeta-partner/structs"
)

var errDB = errors.New("db error")

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
