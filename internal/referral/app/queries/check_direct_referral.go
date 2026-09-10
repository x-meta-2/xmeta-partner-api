package queries

import (
	"errors"

	"xmeta-partner/database"

	"gorm.io/gorm"
)

// CheckDirectReferralHandler verifies the active partner owned by a Cognito UID
// and whether the target Cognito UID has that partner as its current direct referrer.
type CheckDirectReferralHandler struct {
	DB *gorm.DB
}

func (h *CheckDirectReferralHandler) Handle(ownerUID, userID string) (bool, error) {
	var partner database.Partner
	if err := h.DB.Where("user_id = ? AND status = ?", ownerUID, database.PartnerStatusActive).First(&partner).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	var count int64
	if err := h.DB.Model(&database.Referral{}).
		Where("partner_id = ? AND referred_user_id = ? AND ended_at IS NULL", partner.ID, userID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
