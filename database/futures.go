package database

import "time"

type FuturesClosedPosition struct {
	ID            string    `gorm:"primaryKey;column:id" json:"id"`
	SK            string    `gorm:"column:sk" json:"sk"`
	UserID        string    `gorm:"column:user_id" json:"userId"`
	TraderUserID  string    `gorm:"column:trader_user_id" json:"traderUserId"`
	AccountID     string    `gorm:"column:account_id" json:"accountId"`
	Symbol        string    `gorm:"column:symbol" json:"symbol"`
	MarketID      string    `gorm:"column:market_id" json:"marketId"`
	PositionID    string    `gorm:"column:position_id" json:"positionId"`
	OrderID       string    `gorm:"column:order_id" json:"orderId"`
	PositionSide  string    `gorm:"column:position_side" json:"positionSide"`
	CloseReason   string    `gorm:"column:close_reason" json:"closeReason"`
	Quantity      float64   `gorm:"column:quantity" json:"quantity"`
	OpenPrice     float64   `gorm:"column:open_price" json:"openPrice"`
	ClosePrice    float64   `gorm:"column:close_price" json:"closePrice"`
	OpenNotional  float64   `gorm:"column:open_notional" json:"openNotional"`
	CloseNotional float64   `gorm:"column:close_notional" json:"closeNotional"`
	RealizedPnl   float64   `gorm:"column:realized_pnl" json:"realizedPnl"`
	HoldingMs     int64     `gorm:"column:holding_ms" json:"holdingMs"`
	OpenedAt      time.Time `gorm:"column:opened_at" json:"openedAt"`
	ClosedAt      time.Time `gorm:"column:closed_at" json:"closedAt"`
	RecordedAt    time.Time `gorm:"column:recorded_at" json:"recordedAt"`
	UbDay         string    `gorm:"column:ub_day" json:"ubDay"`
	UbMonth       string    `gorm:"column:ub_month" json:"ubMonth"`
}

func (FuturesClosedPosition) TableName() string {
	return "futures_closed_positions"
}
