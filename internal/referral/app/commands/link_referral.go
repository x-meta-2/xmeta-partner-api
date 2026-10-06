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
	hasCycle, err := h.wouldCreateCycle(userID, partner.UserID)
	if err != nil {
		return err
	}
	if hasCycle {
		return domain.ErrCircularReferral
	}

	return h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", userAdvisoryKey(userID)).Error; err != nil {
			return err
		}

		var existing database.Referral
		err := tx.Where("referred_user_id = ? AND ended_at IS NULL", userID).
			First(&existing).Error
		if err == nil {
			if existing.PartnerID != link.PartnerID {
				return domain.ErrActiveReferralExists
			}
			if existing.ReferralLinkID != nil && *existing.ReferralLinkID != link.ID {
				return tx.Model(&existing).UpdateColumn("referral_link_id", link.ID).Error
			}
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		now := time.Now()

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

func (h *LinkReferralHandler) wouldCreateCycle(userID, targetPartnerUserID string) (bool, error) {
	var count int64
	err := h.DB.Raw(`
		WITH RECURSIVE referral_upline AS (
			SELECT r.partner_id, p.user_id, 1 AS depth
			FROM referrals r
			JOIN partners p ON p.id = r.partner_id AND p.deleted_at IS NULL
			WHERE r.referred_user_id = ?
				AND r.ended_at IS NULL
				AND r.deleted_at IS NULL

			UNION ALL

			SELECT r.partner_id, p.user_id, referral_upline.depth + 1
			FROM referral_upline
			JOIN referrals r ON r.referred_user_id = referral_upline.user_id
				AND r.ended_at IS NULL
				AND r.deleted_at IS NULL
			JOIN partners p ON p.id = r.partner_id AND p.deleted_at IS NULL
			WHERE referral_upline.depth < 50
		)
		SELECT COUNT(*)
		FROM referral_upline
		JOIN partners current_partner ON current_partner.id = referral_upline.partner_id
			AND current_partner.deleted_at IS NULL
		WHERE current_partner.user_id = ?
	`, targetPartnerUserID, userID).Scan(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func userAdvisoryKey(userID string) int64 {
	var h uint64 = 14695981039346656037
	for i := 0; i < len(userID); i++ {
		h ^= uint64(userID[i])
		h *= 1099511628211
	}
	return int64(h)
}
