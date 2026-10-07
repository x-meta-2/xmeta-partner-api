package commands

import (
	"errors"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/referral/app/dto"
	"xmeta-partner/internal/referral/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UnlinkReferralHandler struct {
	DB *gorm.DB
}

const userUnlinkCooldown = 7 * 24 * time.Hour

func (h *UnlinkReferralHandler) HandleUser(userID string) (dto.ReferralUnlinkResult, error) {
	var result dto.ReferralUnlinkResult
	now := time.Now()

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", userAdvisoryKey(userID)).Error; err != nil {
			return err
		}

		var referral database.Referral
		err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("referred_user_id = ? AND ended_at IS NULL", userID).
			First(&referral).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNoActiveReferral
		}
		if err != nil {
			return err
		}

		referral.EndedAt = &now
		referral.Status = database.ReferralStatusUnlinked
		if err := tx.Save(&referral).Error; err != nil {
			return err
		}

		result = dto.ReferralUnlinkResult{
			ReferralID:   referral.ID,
			Status:       referral.Status,
			UnlinkedAt:   now,
			NextLinkAt:   now.Add(userUnlinkCooldown),
			NextUnlinkAt: now.Add(userUnlinkCooldown),
		}
		return nil
	})
	if err != nil {
		return dto.ReferralUnlinkResult{}, err
	}
	return result, nil
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
