package structs

type PayoutListParams struct {
	PaginationInput
	Query     string  `json:"query"`
	Status    *string `json:"status"`
	PartnerID *string `json:"partnerId"`
}

type PayoutReviewParams struct {
	FailureReason string `json:"failureReason"`
}

type PayoutCompleteParams struct {
	TransactionID         string                 `json:"transactionId"`
	TransferService       string                 `json:"transferService"`
	TransferRequestID     string                 `json:"transferRequestId"`
	TransferTransactionID string                 `json:"transferTransactionId"`
	TransferResponse      map[string]interface{} `json:"transferResponse"`
}
