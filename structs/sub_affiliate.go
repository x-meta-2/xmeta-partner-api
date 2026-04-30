package structs

type SubAffiliateListParams struct {
	PaginationInput
	Status *string `json:"status"`
}

type SubAffiliateInviteParams struct {
	Email        string   `json:"email" binding:"required"`
	OverrideRate *float64 `json:"overrideRate"`
}

type SubAffiliateStatsParams struct {
	PaginationInput
}
