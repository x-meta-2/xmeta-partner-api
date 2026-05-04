package dto

import "xmeta-partner/database"

type ApplicationListResult struct {
	Items []database.PartnerApplication `json:"items"`
	Total int                           `json:"total"`
}
