package partner

import (
	"net/http"

	"xmeta-partner/controllers/common"
	internalReferral "xmeta-partner/internal/referral"
	"xmeta-partner/middlewares"
	"xmeta-partner/structs"

	"github.com/gin-gonic/gin"
)

type LinkController struct {
	common.Controller
	Service *internalReferral.Service
}

func (co LinkController) Register(router *gin.RouterGroup) {
	co.Service = internalReferral.NewService(co.DB)
	r := router.Use(middlewares.PartnerAuth(co.DB))
	{
		r.POST("/list", co.List)
		r.POST("/create", co.Create)
	}
}

// List
// @Summary       List referral links
// @Description   Returns a paginated list of the partner's referral links
// @Tags          Partner Links
// @Accept        json
// @Produce       json
// @Param         request body structs.ReferralListParams true "Filters and pagination"
// @Success       200 {object} structs.ResponseBody{body=structs.PaginationResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/links/list [post]
func (co LinkController) List(c *gin.Context) {
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

	result, err := co.Service.Queries.ListLinks.Handle(partner.ID, params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// Create
// @Summary       Create referral link
// @Description   Creates a new referral link for the authenticated partner
// @Tags          Partner Links
// @Accept        json
// @Produce       json
// @Param         request body structs.ReferralLinkCreateParams true "Link create payload"
// @Success       200 {object} structs.ResponseBody{body=database.ReferralLink}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/links/create [post]
func (co LinkController) Create(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var params structs.ReferralLinkCreateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Commands.CreateLink.Handle(partner.ID, params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

