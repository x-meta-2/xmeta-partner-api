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

func NextAllowedUnlinkAt(db *gorm.DB, userID string) (*time.Time, error) {
	var referral database.Referral
	err := db.
		Where("referred_user_id = ? AND status = ? AND ended_at IS NOT NULL", userID, database.ReferralStatusUnlinked).
		Order("ended_at desc").
		First(&referral).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if referral.EndedAt == nil {
		return nil, nil
	}
	next := referral.EndedAt.Add(userUnlinkCooldown)
	return &next, nil
}

func CanUnlinkNow(db *gorm.DB, userID string, now time.Time) (bool, *time.Time, error) {
	nextUnlinkAt, err := NextAllowedUnlinkAt(db, userID)
	if err != nil {
		return false, nil, err
	}
	if nextUnlinkAt != nil && nextUnlinkAt.After(now) {
		return false, nextUnlinkAt, nil
	}
	return true, nil, nil
}

func (h *UnlinkReferralHandler) HandleUser(userID string) (dto.ReferralUnlinkResult, error) {
	var result dto.ReferralUnlinkResult
	now := time.Now()

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", userAdvisoryKey(userID)).Error; err != nil {
			return err
		}

		canUnlink, _, err := CanUnlinkNow(tx, userID, now)
		if err != nil {
			return err
		}
		if !canUnlink {
			return domain.ErrUnlinkCooldown
		}

		var referral database.Referral
		err = tx.
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
