package queries

import (
	"xmeta-partner/internal/payout/app/dto"
	"xmeta-partner/internal/payout/port"
)

type AdminPayoutDetailHandler struct {
	Payouts port.PayoutRepo
}

func (h *AdminPayoutDetailHandler) Handle(id string) (*dto.PayoutDetail, error) {
	payout, items, err := h.Payouts.AdminDetail(id)
	if err != nil {
		return nil, err
	}

	return &dto.PayoutDetail{Payout: *payout, Items: items}, nil
}
