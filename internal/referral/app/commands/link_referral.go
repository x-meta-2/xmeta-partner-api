package commands

import (
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

	now := time.Now()

	return h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&database.Referral{}).
			Where("referred_user_id = ? AND ended_at IS NULL", userID).
			Updates(map[string]interface{}{
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
		tx.Model(&database.Referral{}).Where("referred_user_id = ?", userID).Count(&historyCount)
		if historyCount == 1 {
			tx.Model(link).UpdateColumn("registrations", gorm.Expr("registrations + 1"))
			tx.Model(&database.Partner{}).Where("id = ?", link.PartnerID).
				UpdateColumn("total_referrals", gorm.Expr("total_referrals + 1"))
		}

		return nil
	})
}
