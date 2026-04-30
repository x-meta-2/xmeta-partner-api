package common

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"xmeta-partner/structs"
	"xmeta-partner/utils"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type Controller struct {
	DB    *gorm.DB
	Minio *minio.Client
}

// SetBody successfully response
func (co Controller) SetBody(c *gin.Context, body interface{}) {
	c.Set("res_status", http.StatusOK)
	c.Set("res_body", structs.ResponseBody{Message: "", Body: body})
}

// SetError error response
func (co Controller) SetError(c *gin.Context, code int, message string) {
	c.Set("res_status", code)
	c.Set("res_body", structs.ResponseBody{Message: message, Body: nil})
}

// GetBody returns status and body from context
func (co Controller) GetBody(c *gin.Context) (int, interface{}) {
	status, exists := c.Get("res_status")
	if !exists {
		status = http.StatusOK
	}
	body, exists := c.Get("res_body")
	if !exists {
		body = structs.ResponseBody{Message: "No response set", Body: nil}
	}
	return status.(int), body
}

// SortDateFilter applies date filtering based on SortDate input
func SortDateFilter(input *structs.PaginationInput) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if input.SortDate != nil {
			start, errStart := time.Parse(time.DateTime, fmt.Sprintf("%s %s", input.SortDate.StartDay, "00:00:00"))
			end, errEnd := time.Parse(time.DateTime, fmt.Sprintf("%s %s", input.SortDate.EndDay, "23:59:59"))

			if errStart == nil && errEnd == nil {
				db = db.Where("created_at between ? and ?", start, end)
			}
		}
		return db
	}
}

// Paginate table
func Paginate(input *structs.PaginationInput) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if input.Export != nil && *input.Export {
			db = SortDateFilter(input)(db)
			return db
		}

		if input.Current == 0 {
			input.Current = 1
		}
		if input.PageSize <= 0 {
			input.PageSize = 20
		}
		offset := (input.Current - 1) * input.PageSize

		db = SortDateFilter(input)(db)

		return db.Offset(offset).Limit(input.PageSize)
	}
}

// Total is return total from query
func Total(db *gorm.DB) int {
	var total int64
	db.Count(&total)
	return int(total)
}

// Sort is do sort scope
func Sort(input structs.Sorter) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for key, val := range input {
			order := "ASC"
			if strings.HasPrefix(strings.ToLower(val), "desc") {
				order = "DESC"
			}
			db = db.Order(fmt.Sprintf("%s %s", key, order))
		}
		return db
	}
}

// Cursor is do pagination scope
func Cursor(input structs.CursorInput) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("id > ?", input.PreviousID).Order("id ASC").Limit(input.Limit)
	}
}

func Like(db *gorm.DB, field string, value *string) *gorm.DB {
	if utils.IsNil(value) {
		return db
	}
	return db.Where(fmt.Sprintf("LOWER(%s) like LOWER(?)", field), "%"+*value+"%")
}

func OrLike(db *gorm.DB, field string, value *string) *gorm.DB {
	if utils.IsNil(value) {
		return db
	}
	return db.Or(fmt.Sprintf("LOWER(%s) like LOWER(?)", field), "%"+*value+"%")
}

func Equal(db *gorm.DB, field string, value interface{}) *gorm.DB {
	if utils.IsNil(value) {
		return db
	}
	return db.Where(fmt.Sprintf("%s = ?", field), value)
}

func Include(db *gorm.DB, field string, values interface{}) *gorm.DB {
	if utils.IsNil(values) {
		return db
	}
	return db.Where(fmt.Sprintf("%s IN (?)", field), values)
}

func Overlap(db *gorm.DB, field string, values interface{}) *gorm.DB {
	if utils.IsNil(values) {
		return db
	}
	return db.Where(fmt.Sprintf("%s && ?", field), values)
}

func OrEqual(db *gorm.DB, field string, value interface{}) *gorm.DB {
	if utils.IsNil(value) {
		return db
	}
	return db.Or(fmt.Sprintf("%s = ?", field), value)
}

func Greater(db *gorm.DB, field string, value interface{}) *gorm.DB {
	if utils.IsNil(value) {
		return db
	}
	return db.Where(fmt.Sprintf("%s > ?", field), value)
}

func EqualGreater(db *gorm.DB, field string, value interface{}) *gorm.DB {
	if utils.IsNil(value) {
		return db
	}
	return db.Where(fmt.Sprintf("%s >= ?", field), value)
}

func OrEqualGreater(db *gorm.DB, field string, value interface{}) *gorm.DB {
	if utils.IsNil(value) {
		return db
	}
	return db.Or(fmt.Sprintf("%s >= ?", field), value)
}

func Lower(db *gorm.DB, field string, value interface{}) *gorm.DB {
	if utils.IsNil(value) {
		return db
	}
	return db.Where(fmt.Sprintf("%s < ?", field), value)
}

func EqualLower(db *gorm.DB, field string, value interface{}) *gorm.DB {
	if utils.IsNil(value) {
		return db
	}
	return db.Where(fmt.Sprintf("%s <= ?", field), value)
}

func OrEqualLower(db *gorm.DB, field string, value interface{}) *gorm.DB {
	if utils.IsNil(value) {
		return db
	}
	return db.Or(fmt.Sprintf("%s <= ?", field), value)
}

func DateEqual(db *gorm.DB, field string, value interface{}) *gorm.DB {
	if utils.IsNil(value) {
		return db
	}
	return db.Where(fmt.Sprintf("DATE(%s) = DATE(?)", field), value)
}

func OrDateEqual(db *gorm.DB, field string, value interface{}) *gorm.DB {
	if utils.IsNil(value) {
		return db
	}
	return db.Or(fmt.Sprintf("DATE(%s) = DATE(?)", field), value)
}

func Between(db *gorm.DB, field string, min interface{}, max interface{}) *gorm.DB {
	if utils.IsNil(min) || utils.IsNil(max) {
		return db
	}
	return db.Where(fmt.Sprintf("%s >= ? AND %s <= ?", field, field), min, max)
}

func OrBetween(db *gorm.DB, field string, min interface{}, max interface{}) *gorm.DB {
	if utils.IsNil(min) || utils.IsNil(max) {
		return db
	}
	return db.Or(fmt.Sprintf("%s >= ? AND %s <= ?", field, field), min, max)
}
