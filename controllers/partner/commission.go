package partner

import (
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
	r := router.Use(middlewares.PartnerAuth(co.DB))
	{
		r.POST("/list", co.List)
		r.POST("/breakdown", co.Breakdown)
		r.POST("/daily-summary", co.DailySummary)
	}
}

// List
// @Summary       List partner commissions
// @Description   Returns a paginated list of commission records for the partner
// @Tags          Partner Commissions
// @Accept        json
// @Produce       json
// @Param         request body structs.CommissionListParams true "Filters and pagination"
// @Success       200 {object} structs.ResponseBody{body=structs.PaginationResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/commissions/list [post]
func (co CommissionController) List(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var params structs.CommissionListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Queries.ListCommissions.Handle(partner.ID, params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// Breakdown
// @Summary       Commission breakdown
// @Description   Returns commissions grouped by source (direct, sub-affiliate, etc.)
// @Tags          Partner Commissions
// @Accept        json
// @Produce       json
// @Param         request body structs.CommissionBreakdownParams true "Breakdown filters"
// @Success       200 {object} structs.ResponseBody
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/commissions/breakdown [post]
func (co CommissionController) Breakdown(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var params structs.CommissionBreakdownParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Queries.CommissionBreakdown.Handle(partner.ID, params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// DailySummary
// @Summary       Daily commission summary
// @Description   Returns daily commission totals for the chart on the earnings page
// @Tags          Partner Commissions
// @Accept        json
// @Produce       json
// @Param         request body structs.ChartParams true "Chart range filter"
// @Success       200 {object} structs.ResponseBody
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/commissions/daily-summary [post]
func (co CommissionController) DailySummary(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var params structs.ChartParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Queries.DailySummary.Handle(partner.ID, params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}
