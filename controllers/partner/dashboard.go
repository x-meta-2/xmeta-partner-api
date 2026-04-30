package partner

import (
	"net/http"

	"xmeta-partner/controllers/common"
	"xmeta-partner/middlewares"
	"xmeta-partner/services"
	partnersvc "xmeta-partner/services/partner"
	"xmeta-partner/structs"

	"github.com/gin-gonic/gin"
)

// DashboardController handles partner dashboard data
type DashboardController struct {
	common.Controller
	Service *partnersvc.DashboardService
}

func (co DashboardController) Register(router *gin.RouterGroup) {
	co.Service = &partnersvc.DashboardService{
		BaseService: services.BaseService{DB: co.DB},
	}
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

	result, err := co.Service.GetSummary(partner.ID, params)
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

	result, err := co.Service.GetEarningsChart(partner.ID, params)
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

	result, err := co.Service.GetReferralChart(partner.ID, params)
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

	result, err := co.Service.GetTierProgress(partner.ID)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}
