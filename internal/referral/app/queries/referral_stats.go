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
	if err := base().Count(&stats.Total).Error; err != nil {
		return stats, err
	}
	if err := base().Where("status = ?", database.ReferralStatusRegistered).Count(&stats.Registered).Error; err != nil {
		return stats, err
	}
	if err := base().Where("status = ?", database.ReferralStatusActive).Count(&stats.Active).Error; err != nil {
		return stats, err
	}
	if err := base().Where("status = ?", database.ReferralStatusInactive).Count(&stats.Inactive).Error; err != nil {
		return stats, err
	}

	return stats, nil
}
