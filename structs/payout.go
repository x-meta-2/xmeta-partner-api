package structs

type PayoutListParams struct {
	PaginationInput
	Status    *string `json:"status"`
	PartnerID *string `json:"partnerId"`
}

type PayoutReviewParams struct {
	FailureReason string `json:"failureReason"`
}

type PayoutCompleteParams struct {
	TransactionID string `json:"transactionId" binding:"required"`
}
