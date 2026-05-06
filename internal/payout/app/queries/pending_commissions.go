package queries

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/payout/app/dto"

	"gorm.io/gorm"
)

type PendingCommissionsHandler struct {
	DB *gorm.DB
}

func (h *PendingCommissionsHandler) Handle(partnerID string) (*dto.PendingInfo, error) {
	var info dto.PendingInfo

	if err := h.DB.Model(&database.Commission{}).
		Where("partner_id = ? AND status = ?", partnerID, "pending").
		Select("COALESCE(SUM(rebate_amount), 0) as pending_amount, COUNT(*) as pending_count").
		Scan(&info).Error; err != nil {
		return nil, err
	}

	return &info, nil
}
