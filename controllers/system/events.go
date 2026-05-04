package system

import (
	"net/http"

	"xmeta-partner/controllers/common"
	internalCommission "xmeta-partner/internal/commission"
	internalReferral "xmeta-partner/internal/referral"
	"xmeta-partner/middlewares"
	"xmeta-partner/structs"

	"github.com/gin-gonic/gin"
)

type EventsController struct {
	common.Controller
	CommissionService *internalCommission.Service
	ReferralService   *internalReferral.Service
}

func (co EventsController) Register(router *gin.RouterGroup) {
	co.CommissionService = internalCommission.NewService(co.DB)
	co.ReferralService = internalReferral.NewService(co.DB)

	r := router.Use(middlewares.InternalAuth())
	{
		r.POST("/trade-event", co.TradeEvent)
		r.GET("/referral-links/:code", co.LookupReferralLink)
		r.POST("/link-referral", co.LinkReferral)
		r.POST("/unlink-referral", co.UnlinkReferral)
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

	err := co.CommissionService.Commands.ProcessTradeEvent.Handle(params)
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
// @Success       200 {object} structs.ResponseBody{body=dto.ReferralLinkLookup}
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

	result, err := co.ReferralService.Queries.LookupLink.Handle(code)
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

	err := co.ReferralService.Commands.LinkReferral.Handle(params.UserID, params.ReferralCode)
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

	if err := co.ReferralService.Commands.UnlinkReferral.Handle(params.UserID); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	co.SetBody(c, structs.SuccessResponse{Success: true})
}

