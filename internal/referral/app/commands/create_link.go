package commands

import (
	"fmt"

	"xmeta-partner/database"
	"xmeta-partner/internal/referral/domain"
	"xmeta-partner/internal/referral/port"
	"xmeta-partner/structs"
	"xmeta-partner/utils"
)

type CreateLinkHandler struct {
	Links port.ReferralLinkRepo
}

func (h *CreateLinkHandler) Handle(partnerID string, params structs.ReferralLinkCreateParams) (database.ReferralLink, error) {
	count, err := h.Links.CountByPartner(partnerID)
	if err != nil {
		return database.ReferralLink{}, err
	}
	if count >= domain.MaxReferralLinksPerPartner {
		return database.ReferralLink{}, domain.ErrMaxLinksReached
	}

	var code string
	if params.Code != "" {
		if err := utils.ValidateReferralCode(params.Code); err != nil {
			return database.ReferralLink{}, err
		}
		taken, err := h.Links.CodeTaken(params.Code)
		if err != nil {
			return database.ReferralLink{}, err
		}
		if taken {
			return database.ReferralLink{}, domain.ErrCodeTaken
		}
		code = params.Code
	} else {
		for attempt := 0; attempt < 5; attempt++ {
			generated, err := utils.GenerateReferralCode()
			if err != nil {
				return database.ReferralLink{}, err
			}
			taken, err := h.Links.CodeTaken(generated)
			if err != nil {
				return database.ReferralLink{}, err
			}
			if !taken {
				code = generated
				break
			}
		}
		if code == "" {
			return database.ReferralLink{}, domain.ErrCodeGenerationFailed
		}
	}

	url := params.URL
	if url == "" {
		url = fmt.Sprintf("https://x-meta.com/?ref=%s", code)
	}

	link := database.ReferralLink{
		PartnerID: partnerID,
		Code:      code,
		URL:       url,
		IsActive:  true,
	}

	if err := h.Links.Create(&link); err != nil {
		return database.ReferralLink{}, err
	}

	return link, nil
}
