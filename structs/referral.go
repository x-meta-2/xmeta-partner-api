package structs

type ReferralListParams struct {
	PaginationInput
	Status *string `json:"status"`
	Query  string  `json:"query"`
}

type ReferralLinkCreateParams struct {
	URL string `json:"url"`
	// Optional partner-supplied code. 5–7 uppercase A-Z/0-9 characters.
	// Empty/omitted → server auto-generates a 7-character code.
	Code string `json:"code"`
}
