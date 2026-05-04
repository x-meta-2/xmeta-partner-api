package queries

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/referral/app/dto"
	"xmeta-partner/internal/referral/port"
	"xmeta-partner/structs"
	"xmeta-partner/utils"
)

type ListReferralsHandler struct {
	Referrals port.ReferralRepo
}

func (h *ListReferralsHandler) Handle(partnerID string, params structs.ReferralListParams) (structs.PaginationResponse, error) {
	referrals, total, err := h.Referrals.List(partnerID, params)
	if err != nil {
		return structs.PaginationResponse{}, err
	}

	items := make([]dto.ReferralListItem, len(referrals))
	for i := range referrals {
		items[i] = toReferralListItem(referrals[i])
	}
	return structs.PaginationResponse{Total: total, Items: items}, nil
}

func toReferralListItem(r database.Referral) dto.ReferralListItem {
	item := dto.ReferralListItem{
		ID:             r.ID,
		PartnerID:      r.PartnerID,
		ReferredUserID: r.ReferredUserID,
		ReferralLinkID: r.ReferralLinkID,
		Status:         r.Status,
		StartedAt:      r.StartedAt,
		EndedAt:        r.EndedAt,
		RegisteredAt: r.RegisteredAt,
		FirstTradeAt: r.FirstTradeAt,
		CreatedAt:      r.CreatedAt,
	}
	if r.ReferredUser != nil {
		item.ReferredUser = &dto.ReferralUserRef{
			ID:          r.ReferredUser.ID,
			MaskedEmail: utils.MaskEmail(r.ReferredUser.Email),
			FirstName:   r.ReferredUser.FirstName,
			LastInitial: utils.LastInitial(r.ReferredUser.LastName),
			KycLevel:    r.ReferredUser.KycLevel,
		}
	}
	return item
}
