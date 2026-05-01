package partner

import (
	"fmt"

	"xmeta-partner/constants"
	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/services"
	"xmeta-partner/structs"
	"xmeta-partner/utils"
)

type LinkService struct {
	services.BaseService
}

func (s *LinkService) List(partnerID string, params structs.ReferralListParams) (structs.PaginationResponse, error) {
	pInput := services.PreparePaginationInput(params.PaginationInput)
	params.PaginationInput = pInput

	orm := s.DB.Model(&database.ReferralLink{}).Where("partner_id = ?", partnerID)

	total := common.Total(orm.Scopes(common.SortDateFilter(&params.PaginationInput)))

	var links []database.ReferralLink
	if err := orm.
		Order("created_at desc").
		Scopes(common.Paginate(&params.PaginationInput)).
		Find(&links).Error; err != nil {
		return structs.PaginationResponse{}, err
	}

	// Stored `registrations` drifts (bumped on first signup, never decremented on unlink).
	// Recompute live so per-link signup counts stay honest.
	for i := range links {
		var live int64
		s.DB.Model(&database.Referral{}).
			Where("referral_link_id = ? AND ended_at IS NULL", links[i].ID).
			Count(&live)
		links[i].Registrations = int(live)
	}

	return structs.PaginationResponse{Total: total, Items: links}, nil
}

// codeTaken reports whether `code` is taken in either the referral_links or
// partners.referral_code namespace — both share the same code namespace.
func (s *LinkService) codeTaken(code string) (bool, error) {
	var linkCount int64
	if err := s.DB.Model(&database.ReferralLink{}).Where("code = ?", code).Count(&linkCount).Error; err != nil {
		return false, err
	}
	if linkCount > 0 {
		return true, nil
	}

	var partnerCount int64
	if err := s.DB.Model(&database.Partner{}).Where("referral_code = ?", code).Count(&partnerCount).Error; err != nil {
		return false, err
	}
	return partnerCount > 0, nil
}

func (s *LinkService) Create(partnerID string, params structs.ReferralLinkCreateParams) (database.ReferralLink, error) {
	var count int64
	if err := s.DB.Model(&database.ReferralLink{}).Where("partner_id = ?", partnerID).Count(&count).Error; err != nil {
		return database.ReferralLink{}, err
	}
	if count >= constants.MaxReferralLinksPerPartner {
		return database.ReferralLink{}, fmt.Errorf("maximum %d referral links per partner", constants.MaxReferralLinksPerPartner)
	}

	var code string
	if params.Code != "" {
		if err := utils.ValidateReferralCode(params.Code); err != nil {
			return database.ReferralLink{}, err
		}
		if taken, err := s.codeTaken(params.Code); err != nil {
			return database.ReferralLink{}, err
		} else if taken {
			return database.ReferralLink{}, fmt.Errorf("referral code %q is already in use", params.Code)
		}
		code = params.Code
	} else {
		for attempt := 0; attempt < 5; attempt++ {
			generated, err := utils.GenerateReferralCode()
			if err != nil {
				return database.ReferralLink{}, err
			}
			taken, err := s.codeTaken(generated)
			if err != nil {
				return database.ReferralLink{}, err
			}
			if !taken {
				code = generated
				break
			}
		}
		if code == "" {
			return database.ReferralLink{}, fmt.Errorf("could not generate a unique referral code after several attempts; try again")
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

	if err := s.DB.Create(&link).Error; err != nil {
		return database.ReferralLink{}, err
	}

	return link, nil
}

// Referral links are intentionally permanent — no Update/Delete exposed.
// Admins can soft-disable a row via `is_active = false` if needed.
