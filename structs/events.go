package structs

// ReferralCheckParams is the internal server-to-server request used by developer-service.
type ReferralCheckParams struct {
	OwnerUID string `json:"ownerUid" binding:"required"`
	UserID   string `json:"userId" binding:"required"`
}

// ReferralCheckResponse is intentionally limited to a boolean so partner data is never exposed.
type ReferralCheckResponse struct {
	IsReferral bool `json:"isReferral"`
}
