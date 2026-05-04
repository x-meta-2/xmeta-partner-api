package commands

import (
	"time"

	"xmeta-partner/database"
	"xmeta-partner/internal/partner/domain"
	"xmeta-partner/internal/partner/port"
)

type RejectApplicationHandler struct {
	Apps port.ApplicationRepo
}

func (h *RejectApplicationHandler) Handle(applicationID string, adminID string, reason string) (*database.PartnerApplication, error) {
	app, err := h.Apps.FindPendingByID(applicationID)
	if err != nil {
		return nil, domain.ErrApplicationNotFound
	}

	now := time.Now()
	app.Status = database.ApplicationStatusRejected
	app.ReviewedBy = &adminID
	app.ReviewedAt = &now
	app.RejectionReason = reason

	if err := h.Apps.Save(app); err != nil {
		return nil, err
	}

	return app, nil
}
