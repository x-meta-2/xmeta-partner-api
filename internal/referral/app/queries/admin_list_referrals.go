package queries

import (
	"xmeta-partner/internal/referral/port"
	"xmeta-partner/structs"
)

type AdminListReferralsHandler struct {
	Referrals port.ReferralRepo
}

func (h *AdminListReferralsHandler) Handle(params structs.AdminReferralListParams) (structs.PaginationResponse, error) {
	referrals, total, err := h.Referrals.AdminList(params)
	if err != nil {
		return structs.PaginationResponse{}, err
	}
	return structs.PaginationResponse{Total: total, Items: referrals}, nil
}
