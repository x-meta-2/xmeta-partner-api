package commands

import (
	"errors"
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/referral/domain"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

type ApproveUnlinkRequestHandler struct {
	DB *gorm.DB
}

func (h *ApproveUnlinkRequestHandler) Handle(id, adminID string, params structs.ReferralUnlinkRequestReviewParams) (database.ReferralUnlinkRequest, error) {
	var request database.ReferralUnlinkRequest
	now := time.Now()

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", id).First(&request).Error; err != nil {
			return domain.ErrUnlinkRequestNotFound
		}
		if request.Status != database.ReferralUnlinkRequestStatusPending {
			return domain.ErrUnlinkRequestReviewed
		}

		result := tx.Model(&database.Referral{}).
			Where("id = ? AND ended_at IS NULL", request.ReferralID).
			Updates(map[string]any{
				"ended_at": now,
				"status":   database.ReferralStatusUnlinked,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrNoActiveReferral
		}

		request.Status = database.ReferralUnlinkRequestStatusApproved
		request.ReviewedBy = &adminID
		request.AdminNote = params.AdminNote
		request.ReviewedAt = &now

		return tx.Save(&request).Error
	})
	if err != nil {
		return database.ReferralUnlinkRequest{}, err
	}

	if err := h.reload(&request); err != nil {
		return database.ReferralUnlinkRequest{}, err
	}

	return request, nil
}

func (h *ApproveUnlinkRequestHandler) reload(request *database.ReferralUnlinkRequest) error {
	return h.DB.
		Preload("ReferredUser").
		Preload("Partner.User").
		Preload("Reviewer").
		First(request, "id = ?", request.ID).Error
}

type RejectUnlinkRequestHandler struct {
	DB *gorm.DB
}

func (h *RejectUnlinkRequestHandler) Handle(id, adminID string, params structs.ReferralUnlinkRequestReviewParams) (database.ReferralUnlinkRequest, error) {
	var request database.ReferralUnlinkRequest
	now := time.Now()

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("id = ?", id).First(&request).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ErrUnlinkRequestNotFound
		}
		if err != nil {
			return err
		}
		if request.Status != database.ReferralUnlinkRequestStatusPending {
			return domain.ErrUnlinkRequestReviewed
		}

		request.Status = database.ReferralUnlinkRequestStatusRejected
		request.ReviewedBy = &adminID
		request.AdminNote = params.AdminNote
		request.ReviewedAt = &now

		return tx.Save(&request).Error
	})
	if err != nil {
		return database.ReferralUnlinkRequest{}, err
	}

	if err := h.reload(&request); err != nil {
		return database.ReferralUnlinkRequest{}, err
	}

	return request, nil
}

func (h *RejectUnlinkRequestHandler) reload(request *database.ReferralUnlinkRequest) error {
	return h.DB.
		Preload("ReferredUser").
		Preload("Partner.User").
		Preload("Reviewer").
		First(request, "id = ?", request.ID).Error
}
