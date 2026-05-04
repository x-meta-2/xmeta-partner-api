package partner

import (
	"net/http"

	"xmeta-partner/controllers/common"
	internalAnalytics "xmeta-partner/internal/analytics"
	internalAuth "xmeta-partner/internal/auth"
	"xmeta-partner/middlewares"
	"xmeta-partner/structs"

	"github.com/gin-gonic/gin"
)

type DashboardController struct {
	common.Controller
	AnalyticsService *internalAnalytics.Service
	AuthService      *internalAuth.Service
}

func (co DashboardController) Register(router *gin.RouterGroup) {
	co.AnalyticsService = internalAnalytics.NewService(co.DB)
	co.AuthService = internalAuth.NewService(co.DB)
	r := router.Use(middlewares.PartnerAuth(co.DB))
	{
		r.POST("/summary", co.Summary)
		r.POST("/earnings-chart", co.EarningsChart)
		r.POST("/referral-chart", co.ReferralChart)
		r.POST("/tier-progress", co.TierProgress)
	}
}

// Summary
// @Summary       Partner dashboard summary
// @Description   Returns top-level KPIs for the partner dashboard
// @Tags          Partner Dashboard
// @Accept        json
// @Produce       json
// @Param         request body structs.DashboardSummaryParams true "Date range filter"
// @Success       200 {object} structs.ResponseBody
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/dashboard/summary [post]
func (co DashboardController) Summary(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var params structs.DashboardSummaryParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.AnalyticsService.Queries.DashboardSummary.Handle(partner.ID, params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// EarningsChart
// @Summary       Partner earnings chart
// @Description   Returns time-bucketed earnings data for the dashboard chart
// @Tags          Partner Dashboard
// @Accept        json
// @Produce       json
// @Param         request body structs.ChartParams true "Chart range filter"
// @Success       200 {object} structs.ResponseBody
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/dashboard/earnings-chart [post]
func (co DashboardController) EarningsChart(c *gin.Context) {
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

	result, err := co.AnalyticsService.Queries.EarningsChart.Handle(partner.ID, params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// ReferralChart
// @Summary       Partner referral chart
// @Description   Returns time-bucketed referral counts for the dashboard chart
// @Tags          Partner Dashboard
// @Accept        json
// @Produce       json
// @Param         request body structs.ChartParams true "Chart range filter"
// @Success       200 {object} structs.ResponseBody
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/dashboard/referral-chart [post]
func (co DashboardController) ReferralChart(c *gin.Context) {
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

	result, err := co.AnalyticsService.Queries.ReferralChart.Handle(partner.ID, params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// TierProgress
// @Summary       Partner tier progression
// @Description   Returns current tier and progress toward next tier thresholds
// @Tags          Partner Dashboard
// @Accept        json
// @Produce       json
// @Success       200 {object} structs.ResponseBody
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/dashboard/tier-progress [post]
func (co DashboardController) TierProgress(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := co.AuthService.GetTierDetails(partner.ID)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}
