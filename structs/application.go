package structs

type ApplicationListParams struct {
	PaginationInput
	Query  string  `json:"query"`
	Status *string `json:"status"`
}

type ApplicationReviewParams struct {
	RejectionReason string `json:"rejectionReason"`
}
