package dto

type AdminSummary struct {
	TotalPartners       int64   `json:"totalPartners"`
	ActivePartners      int64   `json:"activePartners"`
	PendingApplications int64   `json:"pendingApplications"`
	TotalCommissions    float64 `json:"totalCommissions"`
	TotalVolume         float64 `json:"totalVolume"`
	TotalReferrals      int64   `json:"totalReferrals"`
	PendingPayouts      float64 `json:"pendingPayouts"`
}

type TrendItem struct {
	Date             string  `json:"date"`
	TotalCommission  float64 `json:"totalCommission"`
	TotalVolume      float64 `json:"totalVolume"`
	TransactionCount int64   `json:"transactionCount"`
}

type FunnelResult struct {
	Clicks        int64 `json:"clicks"`
	Registrations int64 `json:"registrations"`
	Traders       int64 `json:"traders"`
}

type DashboardSummary struct {
	TotalEarnings     float64 `json:"totalEarnings"`
	MonthEarnings     float64 `json:"monthEarnings"`
	PendingCommission float64 `json:"pendingCommission"`
	TotalReferrals    int     `json:"totalReferrals"`
	ActiveReferrals   int64   `json:"activeReferrals"`
	TotalVolume       float64 `json:"totalVolume"`
	ConversionRate    float64 `json:"conversionRate"`
}

type EarningsChartItem struct {
	Date        string  `json:"date"`
	Commissions float64 `json:"commissions"`
	TradeVolume float64 `json:"tradeVolume"`
}

type ReferralChartItem struct {
	Date    string `json:"date"`
	Signups int64  `json:"signups"`
}
