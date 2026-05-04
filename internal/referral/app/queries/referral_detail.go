package queries

import (
	"xmeta-partner/internal/referral/app/dto"
	"xmeta-partner/internal/referral/domain"
	"xmeta-partner/internal/referral/port"
)

type ReferralDetailHandler struct {
	Referrals port.ReferralRepo
}

func (h *ReferralDetailHandler) Handle(partnerID, id string) (dto.ReferralListItem, error) {
	referral, err := h.Referrals.FindByIDAndPartner(id, partnerID)
	if err != nil {
		return dto.ReferralListItem{}, domain.ErrReferralNotFound
	}
	return toReferralListItem(*referral), nil
}
