package admin

import (
	"net/http"

	"xmeta-partner/controllers/common"
	"xmeta-partner/middlewares"
	"xmeta-partner/services"
	adminsvc "xmeta-partner/services/admin"
	"xmeta-partner/structs"

	"github.com/gin-gonic/gin"
)

// ReferralController exposes admin views over the partner-referral graph.
type ReferralController struct {
	common.Controller
	Service *adminsvc.ReferralService
}

func (co ReferralController) Register(router *gin.RouterGroup) {
	co.Service = &adminsvc.ReferralService{
		BaseService: services.BaseService{DB: co.DB},
	}
	r := router.Use(middlewares.AdminAuth(co.DB), middlewares.HasPermission("manage_partners"))
	{
		r.POST("/list", co.List)
		r.GET("/detail/:id", co.Detail)
	}
}

// List
// @Summary       List partner referrals (admin)
// @Description   Returns paginated referrals across all partners or filtered to one (`partnerId`). Hides historical (ended_at NOT NULL) rows by default; set `includeHistorical=true` to inspect past partner relationships.
// @Tags          Admin Referrals
// @Accept        json
// @Produce       json
// @Param         request body structs.AdminReferralListParams true "Filters and pagination"
// @Success       200 {object} structs.ResponseBody{body=structs.PaginationResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/referrals/list [post]
func (co ReferralController) List(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.AdminReferralListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.List(params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// Detail
// @Summary       Referral detail (admin)
// @Description   Returns a referral with the linked user + partner, plus the user's commission history under that partner (last 50) and aggregate totals.
// @Tags          Admin Referrals
// @Accept        json
// @Produce       json
// @Param         id path string true "Referral ID"
// @Success       200 {object} structs.ResponseBody
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/referrals/detail/{id} [get]
func (co ReferralController) Detail(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	id := c.Param("id")
	if id == "" {
		co.SetError(c, http.StatusBadRequest, "id is required")
		return
	}

	result, err := co.Service.Detail(id)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}
