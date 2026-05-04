package commands

import (
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/referral/domain"

	"gorm.io/gorm"
)

type UnlinkReferralHandler struct {
	DB *gorm.DB
}

func (h *UnlinkReferralHandler) Handle(userID string) error {
	now := time.Now()
	result := h.DB.Model(&database.Referral{}).
		Where("referred_user_id = ? AND ended_at IS NULL", userID).
		Updates(map[string]interface{}{
			"ended_at": now,
			"status":   database.ReferralStatusUnlinked,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNoActiveReferral
	}
	return nil
}
