package system

import (
	"errors"
	"net/http"

	"xmeta-partner/controllers/common"
	internalReferral "xmeta-partner/internal/referral"
	"xmeta-partner/internal/referral/domain"
	"xmeta-partner/middlewares"
	"xmeta-partner/structs"

	"github.com/gin-gonic/gin"
)

type EventsController struct {
	common.Controller
	ReferralService *internalReferral.Service
}

func (co EventsController) Register(router *gin.RouterGroup) {
	co.ReferralService = internalReferral.NewService(co.DB)

	r := router.Use(middlewares.InternalAuth())
	{
		r.GET("/referral-links/:code", co.LookupReferralLink)
		r.POST("/link-referral", co.LinkReferral)
		r.POST("/unlink-referral", co.UnlinkReferral)
		r.GET("/referrals/current/:userId", co.CurrentReferral)
		r.POST("/referral-unlink-requests", co.CreateUnlinkRequest)
		r.POST("/referrals/check-user", co.CheckDirectReferral)
	}
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
// @Description   Called by xmeta-monorepo right after Cognito signup completes for any user that arrived through a `?ref=CODE` link. Mirrors `POST /partner/auth/link-referral` but is server-to-server (X-Internal-API-Key) instead of user-authenticated.
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
		switch {
		case errors.Is(err, domain.ErrLinkNotFound):
			co.SetError(c, http.StatusNotFound, err.Error())
		case errors.Is(err, domain.ErrPartnerNotActive),
			errors.Is(err, domain.ErrSelfReferral),
			errors.Is(err, domain.ErrActiveReferralExists):
			co.SetError(c, http.StatusBadRequest, err.Error())
		default:
			co.SetError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	co.SetBody(c, structs.SuccessResponse{Success: true})
}

// UnlinkReferral
// @Summary       Unlink a user from their current partner
// @Description   Server-side counterpart to `POST /partner/auth/unlink-referral`. Called by xmeta-monorepo on account closure, compliance flags, or other system-driven detachments (X-Internal-API-Key). Past commissions stay attributed to whoever was active at trade time.
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

// CurrentReferral
// @Summary       Get a user's current partner referral
// @Description   Internal-only lookup. Returns the user's active partner referral or null when the user is not currently linked to a partner.
// @Tags          System Events
// @Produce       json
// @Param         userId path string true "Cognito user ID"
// @Success       200 {object} structs.ResponseBody
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      InternalKey
// @Router        /internal/referrals/current/{userId} [get]
func (co EventsController) CurrentReferral(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	userID := c.Param("userId")
	if userID == "" {
		co.SetError(c, http.StatusBadRequest, "userId is required")
		return
	}

	result, err := co.ReferralService.Queries.CurrentReferral.Handle(userID)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, "Could not get current referral")
		return
	}

	co.SetBody(c, result)
}

// CreateUnlinkRequest
// @Summary       Create a user unlink request
// @Description   Internal-only endpoint used by account-service when a logged-in user asks to stop following their current partner. This only creates a pending request; admin approval performs the actual unlink.
// @Tags          System Events
// @Accept        json
// @Produce       json
// @Param         request body structs.ReferralUnlinkRequestCreateParams true "Unlink request payload"
// @Success       200 {object} structs.ResponseBody{body=database.ReferralUnlinkRequest}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      InternalKey
// @Router        /internal/referral-unlink-requests [post]
func (co EventsController) CreateUnlinkRequest(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.ReferralUnlinkRequestCreateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.ReferralService.Commands.CreateUnlinkRequest.Handle(params)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNoActiveReferral),
			errors.Is(err, domain.ErrPendingUnlinkRequest):
			co.SetError(c, http.StatusBadRequest, err.Error())
		default:
			co.SetError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	co.SetBody(c, result)
}

// CheckDirectReferral
// @Summary       Check an active partner's direct referral
// @Description   Internal-only lookup. OwnerUID and UserID are Cognito user IDs. Non-partners, unknown users, and referrals owned by another partner return false.
// @Tags          System Events
// @Accept        json
// @Produce       json
// @Param         request body structs.ReferralCheckParams true "Referral check payload"
// @Success       200 {object} structs.ResponseBody{body=structs.ReferralCheckResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      InternalKey
// @Router        /internal/referrals/check-user [post]
func (co EventsController) CheckDirectReferral(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.ReferralCheckParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	isReferral, err := co.ReferralService.Queries.CheckDirectReferral.Handle(params.OwnerUID, params.UserID)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, "Could not check referral")
		return
	}
	co.SetBody(c, structs.ReferralCheckResponse{IsReferral: isReferral})
}
