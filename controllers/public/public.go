package public

import (
	"net/http"

	"xmeta-partner/controllers/common"
	"xmeta-partner/database"
	"xmeta-partner/services"
	partnersvc "xmeta-partner/services/partner"

	"github.com/gin-gonic/gin"
)

// Controller — public endpoints requiring no authentication.
// Used by the marketing landing page (tier table) and referral link clicks.
type Controller struct {
	common.Controller
	Service *partnersvc.AuthService
}

func (co Controller) Register(router *gin.RouterGroup) {
	co.Service = &partnersvc.AuthService{
		BaseService: services.BaseService{DB: co.DB},
	}
	router.GET("/ref/:code", co.TrackClick)
	router.GET("/tiers", co.ListTiers)
}

// TrackClick
// @Summary       Track referral link click
// @Description   Records a click on a referral code and returns redirect info
// @Tags          Public
// @Accept        json
// @Produce       json
// @Param         code path string true "Referral code"
// @Success       200 {object} structs.ResponseBody
// @Failure       400 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Router        /public/partner/ref/{code} [get]
func (co Controller) TrackClick(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	code := c.Param("code")
	if code == "" {
		co.SetError(c, http.StatusBadRequest, "referral code is required")
		return
	}

	result, err := co.Service.TrackClick(code, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// ListTiers exposes commission tiers for unauthenticated visitors (landing
// page table). Read-only — admin tier mutations live under /admin/partner.
// @Summary       List commission tiers
// @Description   Returns all commission tiers ordered by level for the public landing page
// @Tags          Public
// @Accept        json
// @Produce       json
// @Success       200 {object} structs.ResponseBody{body=[]database.PartnerTier}
// @Failure       500 {object} structs.ErrorResponse
// @Router        /public/partner/tiers [get]
func (co Controller) ListTiers(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var tiers []database.PartnerTier
	if err := co.DB.Order("level asc").Find(&tiers).Error; err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, tiers)
}
