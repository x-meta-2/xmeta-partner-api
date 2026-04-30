package services

import (
	"xmeta-partner/controllers/common"
	"xmeta-partner/structs"

	"gorm.io/gorm"
)

// BaseService provides shared database access for all services
type BaseService struct {
	DB *gorm.DB
}

// PreparePaginationInput ensures pagination defaults
func PreparePaginationInput(input structs.PaginationInput) structs.PaginationInput {
	if input.Current <= 0 && input.Page > 0 {
		input.Current = input.Page
	}
	if input.Current <= 0 {
		input.Current = 1
	}
	if input.PageSize <= 0 && input.Limit > 0 {
		input.PageSize = input.Limit
	}
	if input.PageSize <= 0 {
		input.PageSize = 20
	}
	return input
}

// PaginateAndCount applies pagination and returns total count with results
func PaginateAndCount[T any](db *gorm.DB, input *structs.PaginationInput) ([]T, int, error) {
	pInput := PreparePaginationInput(*input)
	*input = pInput

	ormWithDateFilter := db.Scopes(common.SortDateFilter(input))
	total := common.Total(ormWithDateFilter)

	var items []T
	if err := db.
		Order("created_at desc").
		Scopes(common.Paginate(input)).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
