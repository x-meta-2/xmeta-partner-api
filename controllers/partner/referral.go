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

// ReferralController handles partner referral management
type ReferralController struct {
	common.Controller
	Service *partnersvc.ReferralService
}

func (co ReferralController) Register(router *gin.RouterGroup) {
	co.Service = &partnersvc.ReferralService{
		BaseService: services.BaseService{DB: co.DB},
	}
	r := router.Use(middlewares.PartnerAuth(co.DB))
	{
		r.POST("/list", co.List)
		r.GET("/detail/:id", co.Detail)
		r.GET("/stats", co.Stats)
	}
}

// List
// @Summary       List partner referrals
// @Description   Returns a paginated list of referrals for the authenticated partner. User PII is masked — only first name + last initial + masked email are exposed.
// @Tags          Partner Referrals
// @Accept        json
// @Produce       json
// @Param         request body structs.ReferralListParams true "Filters and pagination"
// @Success       200 {object} structs.ResponseBody{body=structs.PaginationResponse{items=[]partner.ReferralListItem}}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/referrals/list [post]
func (co ReferralController) List(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var params structs.ReferralListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.List(partner.ID, params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// Detail
// @Summary       Get referral detail
// @Description   Returns one referral for the authenticated partner. User PII is masked — only first name + last initial + masked email are exposed.
// @Tags          Partner Referrals
// @Accept        json
// @Produce       json
// @Param         id path string true "Referral ID"
// @Success       200 {object} structs.ResponseBody{body=partner.ReferralListItem}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/referrals/detail/{id} [get]
func (co ReferralController) Detail(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := c.Param("id")
	if id == "" {
		co.SetError(c, http.StatusBadRequest, "id is required")
		return
	}

	result, err := co.Service.Detail(partner.ID, id)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// Stats
// @Summary       Referral aggregate stats
// @Description   Returns aggregate referral counts and conversion stats for the partner
// @Tags          Partner Referrals
// @Accept        json
// @Produce       json
// @Success       200 {object} structs.ResponseBody
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/referrals/stats [get]
func (co ReferralController) Stats(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := co.Service.Stats(partner.ID)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}
