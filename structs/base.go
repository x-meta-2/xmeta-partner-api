package structs

type (
	// Response struct
	Response struct {
		StatusCode int
		Body       ResponseBody
	}

	// ResponseBody struct
	ResponseBody struct {
		Message string      `json:"message"`
		Body    interface{} `json:"body"`
	}

	Sorter map[string]string

	// PaginationInput struct
	PaginationInput struct {
		PageSize int       `json:"pageSize" form:"pageSize"`
		Current  int       `json:"current" form:"current"`
		Page     int       `json:"page" form:"page"`     // alias for Current
		Limit    int       `json:"limit" form:"limit"`   // alias for PageSize
		Export   *bool     `json:"export" form:"export"`
		SortDate *SortDate `json:"sortDate" form:"sortDate"`
	}

	SortDate struct {
		StartDay string `json:"start_day" form:"start_day"`
		EndDay   string `json:"end_day" form:"end_day"`
	}

	// PaginationResponse
	PaginationResponse struct {
		Total int         `json:"total"`
		Items interface{} `json:"items"`
	}

	// CursorInput
	CursorInput struct {
		Limit      int  `form:"limit" json:"limit"`
		PreviousID uint `form:"previous_id" json:"previous_id"`
		HasAll     bool `form:"has_all" json:"has_all"`
	}

	// CursorResponse
	CursorResponse struct {
		HasNext bool        `json:"has_next"`
		Items   interface{} `json:"items"`
	}

	// SuccessResponse
	SuccessResponse struct {
		Success bool `json:"success"`
	}

	SumStruct struct {
		Amount float64 `json:"amount"`
		Name   string  `json:"name"`
	}

	CountStruct struct {
		Count float64 `json:"count"`
	}

	// FileInput struct
	FileInput struct {
		ID           int     `json:"id"`
		FileName     string  `json:"file_name"`
		OriginalName string  `json:"original_name"`
		PhysicalPath string  `json:"physical_path"`
		Extention    string  `json:"extention"`
		FileSize     float64 `json:"file_size"`
	}

	ErrorResponse struct {
		StatusCode int    `json:"status_code"`
		ErrorMsg   string `json:"error_msg"`
		Body       string `json:"body"`
	}

	ListResponse struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
	}
)
