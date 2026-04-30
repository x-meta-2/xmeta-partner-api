package system

import (
	"net/http"

	"xmeta-partner/controllers/common"
	"xmeta-partner/middlewares"
	"xmeta-partner/services"
	partnersvc "xmeta-partner/services/partner"
	"xmeta-partner/structs"

	"github.com/gin-gonic/gin"
)

// EventsController handles internal API calls from monorepo
type EventsController struct {
	common.Controller
	CommissionService *services.CommissionEngineService
	ReferralService   *partnersvc.ReferralService
}

func (co EventsController) Register(router *gin.RouterGroup) {
	co.CommissionService = &services.CommissionEngineService{
		BaseService: services.BaseService{DB: co.DB},
	}
	co.ReferralService = &partnersvc.ReferralService{
		BaseService: services.BaseService{DB: co.DB},
	}

	r := router.Use(middlewares.InternalAuth())
	{
		r.POST("/trade-event", co.TradeEvent)
		r.GET("/referral-links/:code", co.LookupReferralLink)
		r.POST("/link-referral", co.LinkReferral)
		r.POST("/unlink-referral", co.UnlinkReferral)
		r.POST("/user-deposited", co.UserDeposited)
	}
}

// TradeEvent
// @Summary       Process a trade event
// @Description   Ingests a trade event from monorepo and runs the commission engine
// @Tags          System Events
// @Accept        json
// @Produce       json
// @Param         request body structs.TradeEventParams true "Trade event payload"
// @Success       200 {object} structs.ResponseBody{body=structs.SuccessResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      InternalKey
// @Router        /internal/trade-event [post]
func (co EventsController) TradeEvent(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.TradeEventParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := co.CommissionService.ProcessTradeEvent(params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, structs.SuccessResponse{Success: true})
}

// LookupReferralLink
// @Summary       Validate a referral code
// @Description   Looks up a referral code so the monorepo signup/settings form can show real-time "valid / belongs to {Name}" feedback before the user submits. Returns 404 if the code is unknown.
// @Tags          System Events
// @Produce       json
// @Param         code path string true "Referral code, e.g. ABC1234"
// @Success       200 {object} structs.ResponseBody{body=partner.ReferralLinkLookup}
// @Failure       401 {object} structs.ErrorResponse
// @Failure       404 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      InternalKey
// @Router        /internal/referral-links/{code} [get]
func (co EventsController) LookupReferralLink(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	code := c.Param("code")
	if code == "" {
		co.SetError(c, http.StatusBadRequest, "code is required")
		return
	}

	result, err := co.ReferralService.LookupReferralLink(code)
	if err != nil {
		co.SetError(c, http.StatusNotFound, err.Error())
		return
	}

	co.SetBody(c, result)
}

// LinkReferral
// @Summary       Link a freshly signed-up user to a partner
// @Description   Called by xmeta-monorepo right after Cognito signup completes for any user that arrived through a `?ref=CODE` link. Mirrors `POST /partner/auth/link-referral` but is server-to-server (X-Internal-Key) instead of user-authenticated.
// @Tags          System Events
// @Accept        json
// @Produce       json
// @Param         request body structs.UserRegisteredParams true "Referral attach payload"
// @Success       200 {object} structs.ResponseBody{body=structs.SuccessResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      InternalKey
// @Router        /internal/link-referral [post]
func (co EventsController) LinkReferral(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.UserRegisteredParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := co.ReferralService.ProcessUserRegistered(params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, structs.SuccessResponse{Success: true})
}

// UnlinkReferral
// @Summary       Unlink a user from their current partner
// @Description   Server-side counterpart to `POST /partner/auth/unlink-referral`. Called by xmeta-monorepo on account closure, compliance flags, or other system-driven detachments. Past commissions stay attributed to whoever was active at trade time.
// @Tags          System Events
// @Accept        json
// @Produce       json
// @Param         request body structs.UnlinkReferralEventParams true "Unlink payload"
// @Success       200 {object} structs.ResponseBody{body=structs.SuccessResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      InternalKey
// @Router        /internal/unlink-referral [post]
func (co EventsController) UnlinkReferral(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.UnlinkReferralEventParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := co.ReferralService.UnlinkReferral(params.UserID); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	co.SetBody(c, structs.SuccessResponse{Success: true})
}

// UserDeposited
// @Summary       Process user deposit
// @Description   Records a referred user deposit and runs deposit-based commission rules
// @Tags          System Events
// @Accept        json
// @Produce       json
// @Param         request body structs.UserDepositedParams true "User deposit payload"
// @Success       200 {object} structs.ResponseBody{body=structs.SuccessResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      InternalKey
// @Router        /internal/user-deposited [post]
func (co EventsController) UserDeposited(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.UserDepositedParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	err := co.ReferralService.ProcessUserDeposited(params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, structs.SuccessResponse{Success: true})
}
