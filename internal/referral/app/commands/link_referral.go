package commands

import (
	"errors"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/referral/domain"
	"xmeta-partner/internal/referral/port"

	"gorm.io/gorm"
)

type LinkReferralHandler struct {
	DB    *gorm.DB
	Links port.ReferralLinkRepo
}

func (h *LinkReferralHandler) Handle(userID, code string) error {
	link, err := h.Links.FindByCodeActive(code)
	if err != nil {
		return domain.ErrLinkNotFound
	}

	var partner database.Partner
	if err := h.DB.Where("id = ?", link.PartnerID).First(&partner).Error; err != nil {
		return domain.ErrLinkNotFound
	}
	if partner.Status != database.PartnerStatusActive {
		return domain.ErrPartnerNotActive
	}
	if partner.UserID == userID {
		return domain.ErrSelfReferral
	}

	return h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", userAdvisoryKey(userID)).Error; err != nil {
			return err
		}

		var existing database.Referral
		err := tx.Where("referred_user_id = ? AND partner_id = ? AND ended_at IS NULL", userID, link.PartnerID).
			First(&existing).Error
		if err == nil {
			if existing.ReferralLinkID != nil && *existing.ReferralLinkID != link.ID {
				return tx.Model(&existing).UpdateColumn("referral_link_id", link.ID).Error
			}
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		now := time.Now()

		if err := tx.Model(&database.Referral{}).
			Where("referred_user_id = ? AND ended_at IS NULL", userID).
			Updates(map[string]any{
				"ended_at": now,
				"status":   database.ReferralStatusUnlinked,
			}).Error; err != nil {
			return err
		}

		referral := database.Referral{
			PartnerID:      link.PartnerID,
			ReferredUserID: userID,
			ReferralLinkID: &link.ID,
			Status:         database.ReferralStatusRegistered,
			StartedAt:      now,
			RegisteredAt:   now,
		}
		if err := tx.Create(&referral).Error; err != nil {
			return err
		}

		var historyCount int64
		if err := tx.Model(&database.Referral{}).Where("referred_user_id = ?", userID).Count(&historyCount).Error; err != nil {
			return err
		}
		if historyCount == 1 {
			if err := tx.Model(link).UpdateColumn("registrations", gorm.Expr("registrations + 1")).Error; err != nil {
				return err
			}
			if err := tx.Model(&database.Partner{}).Where("id = ?", link.PartnerID).
				UpdateColumn("total_referrals", gorm.Expr("total_referrals + 1")).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func userAdvisoryKey(userID string) int64 {
	var h uint64 = 14695981039346656037
	for i := 0; i < len(userID); i++ {
		h ^= uint64(userID[i])
		h *= 1099511628211
	}
	return int64(h)
}
