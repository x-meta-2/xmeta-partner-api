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
	latestReferralIDs := h.DB.
		Table("referrals").
		Select("DISTINCT ON (referred_user_id) id").
		Where("partner_id = ? AND deleted_at IS NULL", partnerID).
		Order("referred_user_id, created_at DESC, id DESC")

	base := func() *gorm.DB {
		return h.DB.Model(&database.Referral{}).
			Where("partner_id = ?", partnerID).
			Where("id IN (?)", latestReferralIDs)
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
	if err := base().Where("status = ?", database.ReferralStatusUnlinked).Count(&stats.Unlinked).Error; err != nil {
		return stats, err
	}

	return stats, nil
}
