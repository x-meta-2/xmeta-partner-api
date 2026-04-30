package structs

import "time"

type DashboardSummaryParams struct {
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
}

type ChartParams struct {
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
	GroupBy   string     `json:"groupBy"` // "day", "week", "month"
}
