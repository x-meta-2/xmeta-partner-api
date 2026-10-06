package queries

import (
	"errors"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/referral/app/dto"

	"gorm.io/gorm"
)

type CurrentReferralHandler struct {
	DB *gorm.DB
}

func (h *CurrentReferralHandler) Handle(userID string) (*dto.CurrentReferral, error) {
	var referral database.Referral
	err := h.DB.
		Preload("Partner.User").
		Preload("ReferralLink").
		Where("referred_user_id = ? AND ended_at IS NULL", userID).
		First(&referral).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	referralCode := ""
	if referral.ReferralLink != nil {
		referralCode = referral.ReferralLink.Code
	} else if referral.Partner != nil {
		referralCode = referral.Partner.ReferralCode
	}

	result := &dto.CurrentReferral{
		ReferralCode: referralCode,
	}

	if referral.Partner != nil {
		result.Partner = dto.CurrentReferralPartner{
			ReferralCode: referral.Partner.ReferralCode,
		}
		if referral.Partner.User != nil {
			result.Partner.Email = referral.Partner.User.Email
			result.Partner.FirstName = referral.Partner.User.FirstName
		}
	}

	canUnlink, nextUnlinkAt, err := canUnlinkNow(h.DB, userID, time.Now())
	if err != nil {
		return nil, err
	}
	result.CanUnlink = canUnlink
	result.NextUnlinkAt = nextUnlinkAt

	return result, nil
}

const userUnlinkCooldown = 7 * 24 * time.Hour

func canUnlinkNow(db *gorm.DB, userID string, now time.Time) (bool, *time.Time, error) {
	var referral database.Referral
	err := db.
		Where("referred_user_id = ? AND status = ? AND ended_at IS NOT NULL", userID, database.ReferralStatusUnlinked).
		Order("ended_at desc").
		First(&referral).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil, nil
	}
	if err != nil {
		return false, nil, err
	}
	if referral.EndedAt == nil {
		return true, nil, nil
	}

	nextUnlinkAt := referral.EndedAt.Add(userUnlinkCooldown)
	if nextUnlinkAt.After(now) {
		return false, &nextUnlinkAt, nil
	}
	return true, nil, nil
}
