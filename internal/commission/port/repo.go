package port

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/commission/app/dto"
	"xmeta-partner/structs"
)

type CommissionRepo interface {
	List(partnerID string, params structs.CommissionListParams) ([]database.Commission, int, error)
	Breakdown(partnerID string, params structs.CommissionBreakdownParams) (float64, error)
	DailySummary(partnerID string, params structs.ChartParams) ([]dto.DailyItem, error)
}
