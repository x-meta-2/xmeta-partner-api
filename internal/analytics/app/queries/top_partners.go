package queries

import (
	"xmeta-partner/database"
	"xmeta-partner/internal/analytics/port"
	"xmeta-partner/structs"
)

type TopPartnersHandler struct {
	Analytics port.AdminAnalyticsRepo
}

func (h *TopPartnersHandler) Handle(params structs.PaginationInput) ([]database.Partner, error) {
	return h.Analytics.TopPartners(params)
}
