package dto

import "xmeta-partner/database"

type TierListResult struct {
	Items []database.PartnerTier `json:"items"`
}
