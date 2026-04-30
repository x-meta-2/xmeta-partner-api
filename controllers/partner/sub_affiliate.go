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

// SubAffiliateController handles sub-affiliate management
type SubAffiliateController struct {
	common.Controller
	Service *partnersvc.SubAffiliateService
}

func (co SubAffiliateController) Register(router *gin.RouterGroup) {
	co.Service = &partnersvc.SubAffiliateService{
		BaseService: services.BaseService{DB: co.DB},
	}
	r := router.Use(middlewares.PartnerAuth(co.DB))
	{
		r.POST("/list", co.List)
		r.POST("/invite", co.Invite)
		r.GET("/invite/:code", co.InviteDetail)
		r.POST("/stats", co.Stats)
	}
}

// List
// @Summary       List sub-affiliates
// @Description   Returns a paginated list of sub-affiliates under the partner
// @Tags          Partner Sub-Affiliates
// @Accept        json
// @Produce       json
// @Param         request body structs.SubAffiliateListParams true "Filters and pagination"
// @Success       200 {object} structs.ResponseBody{body=structs.PaginationResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/sub-affiliates/list [post]
func (co SubAffiliateController) List(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var params structs.SubAffiliateListParams
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

// Invite
// @Summary       Invite a sub-affiliate
// @Description   Creates a new sub-affiliate invite under the authenticated partner
// @Tags          Partner Sub-Affiliates
// @Accept        json
// @Produce       json
// @Param         request body structs.SubAffiliateInviteParams true "Invite payload"
// @Success       200 {object} structs.ResponseBody{body=database.SubAffiliateInvite}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/sub-affiliates/invite [post]
func (co SubAffiliateController) Invite(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var params structs.SubAffiliateInviteParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Invite(partner.ID, params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// InviteDetail
// @Summary       Get sub-affiliate invite detail
// @Description   Returns the sub-affiliate invite identified by code
// @Tags          Partner Sub-Affiliates
// @Accept        json
// @Produce       json
// @Param         code path string true "Invite code"
// @Success       200 {object} structs.ResponseBody{body=database.SubAffiliateInvite}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/sub-affiliates/invite/{code} [get]
func (co SubAffiliateController) InviteDetail(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	code := c.Param("code")
	if code == "" {
		co.SetError(c, http.StatusBadRequest, "code is required")
		return
	}

	result, err := co.Service.InviteDetail(partner.ID, code)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// Stats
// @Summary       Sub-affiliate stats
// @Description   Returns aggregate statistics for the partner's sub-affiliate network
// @Tags          Partner Sub-Affiliates
// @Accept        json
// @Produce       json
// @Param         request body structs.SubAffiliateStatsParams true "Stats filters"
// @Success       200 {object} structs.ResponseBody
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/sub-affiliates/stats [post]
func (co SubAffiliateController) Stats(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var params structs.SubAffiliateStatsParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Stats(partner.ID, params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}
