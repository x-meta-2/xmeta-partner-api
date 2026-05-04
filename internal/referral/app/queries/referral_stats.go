package queries

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/referral/app/dto"

	"gorm.io/gorm"
)

type ReferralStatsHandler struct {
	DB *gorm.DB
}

func (h *ReferralStatsHandler) Handle(partnerID string) (dto.ReferralStats, error) {
	base := func() *gorm.DB {
		return h.DB.Model(&database.Referral{}).
			Where("partner_id = ? AND ended_at IS NULL", partnerID)
	}

	var stats dto.ReferralStats
	base().Count(&stats.Total)
	base().Where("status = ?", "registered").Count(&stats.Registered)
	base().Where("status = ?", "active").Count(&stats.Active)
	base().Where("status = ?", "inactive").Count(&stats.Inactive)

	return stats, nil
}
