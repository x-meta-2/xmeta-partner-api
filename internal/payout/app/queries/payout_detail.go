package queries

import (
	"xmeta-partner/internal/payout/app/dto"
	"xmeta-partner/internal/payout/port"
)

type PayoutDetailHandler struct {
	Payouts port.PayoutRepo
}

func (h *PayoutDetailHandler) Handle(partnerID, id string) (*dto.PayoutDetail, error) {
	payout, items, err := h.Payouts.Detail(partnerID, id)
	if err != nil {
		return nil, err
	}

	return &dto.PayoutDetail{Payout: *payout, Items: items}, nil
}
