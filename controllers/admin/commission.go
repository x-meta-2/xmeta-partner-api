package admin

import (
	"fmt"
	"net/http"

	"xmeta-partner/controllers/common"
	internalCommission "xmeta-partner/internal/commission"
	"xmeta-partner/middlewares"
	"xmeta-partner/structs"

	"github.com/gin-gonic/gin"
)

type CommissionController struct {
	common.Controller
	Service *internalCommission.Service
}

func (co CommissionController) Register(router *gin.RouterGroup) {
	co.Service = internalCommission.NewService(co.DB)

	view := router.Use(middlewares.AdminAuth(co.DB), middlewares.HasPermission("view_partner_commissions"))
	{
		view.POST("/list", co.List)
	}

	manage := router.Use(middlewares.AdminAuth(co.DB), middlewares.HasPermission("manage_partner_commissions"))
	{
		manage.POST("/import", co.Import)
	}
}

// List
// @Summary       List all commissions (admin)
// @Description   Returns a paginated list of all commissions across all partners
// @Tags          Admin Commissions
// @Accept        json
// @Produce       json
// @Param         request body structs.AdminCommissionListParams true "Filters and pagination"
// @Success       200 {object} structs.ResponseBody{body=structs.PaginationResponse}
// @Router        /admin/partner/commissions/list [post]
func (co CommissionController) List(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.AdminCommissionListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Queries.AdminListCommissions.Handle(params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

type ImportResult struct {
	Total   int           `json:"total"`
	Success int           `json:"success"`
	Skipped int           `json:"skipped"`
	Failed  int           `json:"failed"`
	Errors  []ImportError `json:"errors"`
}

type ImportError struct {
	Row     int    `json:"row"`
	UserID  string `json:"userId"`
	Message string `json:"message"`
}

// Import
// @Summary       Batch import trade events
// @Description   Accepts an array of trade events (from Excel upload) and processes each through the commission engine
// @Tags          Admin Commissions
// @Accept        json
// @Produce       json
// @Param         request body []structs.TradeEventParams true "Array of trade events"
// @Success       200 {object} structs.ResponseBody{body=ImportResult}
// @Router        /admin/partner/commissions/import [post]
func (co CommissionController) Import(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var events []structs.TradeEventParams
	if err := c.ShouldBindJSON(&events); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	if len(events) == 0 {
		co.SetError(c, http.StatusBadRequest, "empty event list")
		return
	}

	result := ImportResult{Total: len(events)}

	for i, event := range events {
		if event.UserID == "" || event.PositionID == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				UserID:  event.UserID,
				Message: "userId and positionId are required",
			})
			continue
		}

		res, err := co.Service.Commands.ProcessTradeEvent.Handle(event)
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Row:     i + 1,
				UserID:  event.UserID,
				Message: fmt.Sprintf("%v", err),
			})
			continue
		}

		if res.Skipped {
			result.Skipped++
			continue
		}

		result.Success++
	}

	co.SetBody(c, result)
}
