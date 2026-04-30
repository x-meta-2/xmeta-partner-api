package structs

type PartnerSignupParams struct {
	// UserID = Cognito sub of the applying user.
	// Controller populates this from the authenticated JWT; client does not send it.
	UserID        string                 `json:"userId"`
	CompanyName   string                 `json:"companyName"`
	Website       string                 `json:"website"`
	SocialMedia   map[string]interface{} `json:"socialMedia"`
	AudienceSize  string                 `json:"audienceSize"`
	PromotionPlan string                 `json:"promotionPlan"`
}

// PartnerProfileUpdateParams — partner-owned fields only.
// Identity fields (firstName/lastName/email) are owned by the users table.
type PartnerProfileUpdateParams struct {
	CompanyName *string                `json:"companyName"`
	Website     *string                `json:"website"`
	SocialMedia map[string]interface{} `json:"socialMedia"`
}


type PartnerListParams struct {
	PaginationInput
	Query  string  `json:"query"`
	Status *string `json:"status"`
	TierID *string `json:"tierId"`
}
