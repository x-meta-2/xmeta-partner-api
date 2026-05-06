package dto

type CommissionBreakdown struct {
	Futures float64 `json:"futures"`
	Total   float64 `json:"total"`
}

type DailyItem struct {
	Date             string  `json:"date"`
	RebateAmount float64 `json:"rebateAmount"`
	TradeVolume      float64 `json:"tradeVolume"`
	Count            int64   `json:"count"`
}
