package queries

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/payout/app/dto"

	"gorm.io/gorm"
)

const MinPayoutAmount = 10.0

type PendingCommissionsHandler struct {
	DB *gorm.DB
}

func (h *PendingCommissionsHandler) Handle(partnerID string) (*dto.PendingInfo, error) {
	info := &dto.PendingInfo{MinPayoutAmount: MinPayoutAmount}

	var pending struct {
		Amount float64
		Count  int64
	}
	if err := h.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND status = ? AND payout_id IS NULL", partnerID, "pending").
		Select("COALESCE(SUM(rebate_amount), 0) as amount, COUNT(*) as count").
		Scan(&pending).Error; err != nil {
		return nil, err
	}
	info.PendingBalance = pending.Amount
	info.PendingCount = pending.Count

	var totalPaid float64
	if err := h.DB.Model(&database.Payout{}).
		Where("partner_id = ? AND status = ?", partnerID, database.PayoutStatusCompleted).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalPaid).Error; err != nil {
		return nil, err
	}
	info.TotalPaid = totalPaid

	var lastPayout database.Payout
	err := h.DB.Where("partner_id = ? AND status = ?", partnerID, database.PayoutStatusCompleted).
		Order("processed_at desc").
		First(&lastPayout).Error
	if err == nil && lastPayout.ProcessedAt != nil {
		d := lastPayout.ProcessedAt.Format("2006-01-02")
		info.LastPayoutDate = &d
	}

	return info, nil
}
