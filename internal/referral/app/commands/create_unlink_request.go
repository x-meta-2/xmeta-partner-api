package commands

import (
	"errors"

	"xmeta-partner/database"
	"xmeta-partner/internal/referral/domain"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type CreateUnlinkRequestHandler struct {
	DB *gorm.DB
}

func (h *CreateUnlinkRequestHandler) Handle(params structs.ReferralUnlinkRequestCreateParams) (database.ReferralUnlinkRequest, error) {
	var result database.ReferralUnlinkRequest

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", userAdvisoryKey(params.UserID)).Error; err != nil {
			return err
		}

		var existing database.ReferralUnlinkRequest
		err := tx.Where("referred_user_id = ? AND status = ?", params.UserID, database.ReferralUnlinkRequestStatusPending).
			First(&existing).Error
		if err == nil {
			return domain.ErrPendingUnlinkRequest
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		var referral database.Referral
		err = tx.
			Preload("ReferralLink").
			Preload("Partner").
			Where("referred_user_id = ? AND ended_at IS NULL", params.UserID).
			First(&referral).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrNoActiveReferral
		}
		if err != nil {
			return err
		}

		referralCode := ""
		if referral.ReferralLink != nil {
			referralCode = referral.ReferralLink.Code
		} else if referral.Partner != nil {
			referralCode = referral.Partner.ReferralCode
		}

		result = database.ReferralUnlinkRequest{
			ReferralID:     referral.ID,
			PartnerID:      referral.PartnerID,
			ReferredUserID: referral.ReferredUserID,
			ReferralCode:   referralCode,
			Reason:         params.Reason,
			Status:         database.ReferralUnlinkRequestStatusPending,
		}

		return tx.Create(&result).Error
	})
	if err != nil {
		return database.ReferralUnlinkRequest{}, err
	}

	return result, nil
}
