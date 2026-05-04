package admin

import (
	"net/http"

	"xmeta-partner/controllers/common"
	internalAnalytics "xmeta-partner/internal/analytics"
	"xmeta-partner/middlewares"
	"xmeta-partner/structs"

	"github.com/gin-gonic/gin"
)

type AnalyticsController struct {
	common.Controller
	Service *internalAnalytics.Service
}

func (co AnalyticsController) Register(router *gin.RouterGroup) {
	co.Service = internalAnalytics.NewService(co.DB)
	r := router.Use(middlewares.AdminAuth(co.DB), middlewares.HasPermission("manage_partners"))
	{
		r.POST("/summary", co.Summary)
		r.POST("/commission-trend", co.CommissionTrend)
		r.POST("/top-partners", co.TopPartners)
		r.POST("/referral-funnel", co.ReferralFunnel)
	}
}

// Summary
// @Summary       Admin analytics summary
// @Description   Returns top-level KPIs for the admin analytics dashboard
// @Tags          Admin Analytics
// @Accept        json
// @Produce       json
// @Param         request body structs.DashboardSummaryParams true "Date range filter"
// @Success       200 {object} structs.ResponseBody
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/analytics/summary [post]
func (co AnalyticsController) Summary(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.DashboardSummaryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Queries.AdminSummary.Handle(params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// CommissionTrend
// @Summary       Commission trend chart
// @Description   Returns time-bucketed commission totals across all partners
// @Tags          Admin Analytics
// @Accept        json
// @Produce       json
// @Param         request body structs.ChartParams true "Chart range filter"
// @Success       200 {object} structs.ResponseBody
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/analytics/commission-trend [post]
func (co AnalyticsController) CommissionTrend(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.ChartParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Queries.CommissionTrend.Handle(params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// TopPartners
// @Summary       Top partners leaderboard
// @Description   Returns the highest-earning partners ranked by commission volume
// @Tags          Admin Analytics
// @Accept        json
// @Produce       json
// @Param         request body structs.PaginationInput true "Pagination input"
// @Success       200 {object} structs.ResponseBody{body=structs.PaginationResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/analytics/top-partners [post]
func (co AnalyticsController) TopPartners(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.PaginationInput
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Queries.TopPartners.Handle(params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// ReferralFunnel
// @Summary       Referral funnel metrics
// @Description   Returns conversion-funnel metrics from click to commissioned trade
// @Tags          Admin Analytics
// @Accept        json
// @Produce       json
// @Param         request body structs.DashboardSummaryParams true "Date range filter"
// @Success       200 {object} structs.ResponseBody
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/analytics/referral-funnel [post]
func (co AnalyticsController) ReferralFunnel(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.DashboardSummaryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Queries.ReferralFunnel.Handle(params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}
