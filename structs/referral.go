package structs

type ReferralListParams struct {
	PaginationInput
	Status *string `json:"status"`
	Query  string  `json:"query"`
}

// AdminReferralListParams — admin variant. Accepts an optional PartnerID
// filter so the same endpoint serves both the global referrals page and
// the per-partner drawer tab. IncludeHistorical defaults to false; the
// admin UI shows currently-linked rows by default and toggles this flag
// when the operator wants to inspect past partner relationships.
type AdminReferralListParams struct {
	PaginationInput
	PartnerID         *string `json:"partnerId"`
	Status            *string `json:"status"`
	Query             string  `json:"query"`
	IncludeHistorical bool    `json:"includeHistorical"`
}

type ReferralLinkCreateParams struct {
	URL string `json:"url"`
	// Optional partner-supplied code. 5–7 uppercase A-Z/0-9 characters.
	// Empty/omitted → server auto-generates a 7-character code.
	Code string `json:"code"`
}
