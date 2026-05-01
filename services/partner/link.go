package partner

import (
	"fmt"

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

	// Override stored `registrations` with a live count of currently
	// linked referrals (ended_at IS NULL). The column is bumped only on
	// the user's first-ever signup and never decremented on unlink, so
	// the stored value drifts. Recomputing keeps the partner-portal
	// per-link signup count honest.
	for i := range links {
		var live int64
		s.DB.Model(&database.Referral{}).
			Where("referral_link_id = ? AND ended_at IS NULL", links[i].ID).
			Count(&live)
		links[i].Registrations = int(live)
	}

	return structs.PaginationResponse{Total: total, Items: links}, nil
}

// MaxReferralLinksPerPartner caps how many links a partner can have. The
// approve flow creates one default link, leaving room for two custom ones.
const MaxReferralLinksPerPartner = 3

// codeTaken reports whether `code` is already used by any referral link or
// any partner's primary referral code — both columns share the same code
// namespace and both have UNIQUE indexes.
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
	if count >= MaxReferralLinksPerPartner {
		return database.ReferralLink{}, fmt.Errorf("maximum %d referral links per partner", MaxReferralLinksPerPartner)
	}

	var code string
	if params.Code != "" {
		if err := utils.ValidateReferralCode(params.Code); err != nil {
			return database.ReferralLink{}, err
		}
		// Custom codes need an explicit uniqueness check so we can return a
		// clean "already in use" message rather than letting the DB unique
		// constraint surface a raw 23505 error to the partner.
		if taken, err := s.codeTaken(params.Code); err != nil {
			return database.ReferralLink{}, err
		} else if taken {
			return database.ReferralLink{}, fmt.Errorf("referral code %q is already in use", params.Code)
		}
		code = params.Code
	} else {
		// Auto-generated codes have a 1-in-78B collision chance; retry up
		// to 5 times against the live `referral_links.code` index before
		// giving up so the approve/create flow is resilient.
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

// Referral links are intentionally permanent — once a partner shares a code
// with their audience, edits/deletes risk breaking inbound referrals and
// muddying audit trails. No Update/Delete service method exposed; admins
// can soft-disable a row in the database (`is_active = false`) if needed.
