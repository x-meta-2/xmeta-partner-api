package queries

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/referral/app/dto"
	"xmeta-partner/internal/referral/domain"

	"gorm.io/gorm"
)

type LookupLinkHandler struct {
	DB *gorm.DB
}

func (h *LookupLinkHandler) Handle(code string) (dto.ReferralLinkLookup, error) {
	var link database.ReferralLink
	if err := h.DB.Where("code = ?", code).First(&link).Error; err != nil {
		return dto.ReferralLinkLookup{}, domain.ErrLinkNotFound
	}

	out := dto.ReferralLinkLookup{
		Code:      link.Code,
		PartnerID: link.PartnerID,
	}

	var partner database.Partner
	if err := h.DB.Preload("User").Where("id = ?", link.PartnerID).First(&partner).Error; err == nil {
		out.IsActive = link.IsActive && partner.Status == database.PartnerStatusActive
		if partner.User != nil {
			out.PartnerEmail = partner.User.Email
			out.PartnerFullName = partner.User.FirstName + " " + partner.User.LastName
		}
	}

	return out, nil
}
