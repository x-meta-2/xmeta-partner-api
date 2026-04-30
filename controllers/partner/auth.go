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

// AuthController — partner auth & onboarding endpoints.
//
// Route grouping by middleware:
//   - [UserAuth]      /status, /apply        — any valid xmeta user
//   - [PartnerAuth]   /info, /profile, /tier — active partners only
//
// Referral link/unlink lives only on the System Events controller
// (`/internal/link-referral`, `/internal/unlink-referral`) because
// regular xmeta users live on xmeta-monorepo, not on the partner portal.
// Monorepo proxies the user action server-to-server with InternalKey.
type AuthController struct {
	common.Controller
	Service *partnersvc.AuthService
}

func (co AuthController) Register(router *gin.RouterGroup) {
	co.Service = &partnersvc.AuthService{
		BaseService: services.BaseService{DB: co.DB},
	}

	// Any authenticated xmeta user (partner эсвэл биш) хандаж болох endpoint-ууд
	userRoutes := router.Group("").Use(middlewares.UserAuth(co.DB))
	{
		userRoutes.GET("/status", co.Status)
		userRoutes.POST("/apply", co.Apply)
	}

	// Зөвхөн active partner хандаж болох endpoint-ууд
	partnerRoutes := router.Group("").Use(middlewares.PartnerAuth(co.DB))
	{
		partnerRoutes.GET("/info", co.Info)
		partnerRoutes.PUT("/profile", co.UpdateProfile)
		partnerRoutes.GET("/tier", co.TierDetails)
	}
}

// Status returns {user, partner, application} — frontend uses this right
// after login to decide which screen (dashboard / pending / apply) to show.
// @Summary       Partner auth status
// @Description   Returns user, partner, and application records for routing decisions
// @Tags          Partner Auth
// @Accept        json
// @Produce       json
// @Success       200 {object} structs.ResponseBody
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/auth/status [get]
func (co AuthController) Status(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	user := middlewares.UserGetAuth(c)
	if user == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := co.Service.GetAuthStatus(user)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// Apply creates a new partner application for the authenticated user.
// @Summary       Submit partner application
// @Description   Creates a partner application for the authenticated user
// @Tags          Partner Auth
// @Accept        json
// @Produce       json
// @Param         request body structs.PartnerSignupParams true "Application payload"
// @Success       200 {object} structs.ResponseBody{body=database.PartnerApplication}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/auth/apply [post]
func (co AuthController) Apply(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	user := middlewares.UserGetAuth(c)
	if user == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var params structs.PartnerSignupParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}
	// Override client-sent userId — trust JWT instead
	params.UserID = user.ID

	result, err := co.Service.Signup(params)
	if err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	co.SetBody(c, result)
}

// Info
// @Summary       Get partner profile
// @Description   Returns the authenticated partner's profile and account info
// @Tags          Partner Auth
// @Accept        json
// @Produce       json
// @Success       200 {object} structs.ResponseBody{body=database.Partner}
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/auth/info [get]
func (co AuthController) Info(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := co.Service.GetInfo(partner.ID)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// UpdateProfile
// @Summary       Update partner profile
// @Description   Updates the authenticated partner's editable profile fields
// @Tags          Partner Auth
// @Accept        json
// @Produce       json
// @Param         request body structs.PartnerProfileUpdateParams true "Profile update payload"
// @Success       200 {object} structs.ResponseBody{body=database.Partner}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/auth/profile [put]
func (co AuthController) UpdateProfile(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var params structs.PartnerProfileUpdateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.UpdateProfile(partner.ID, params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// TierDetails
// @Summary       Get partner tier details
// @Description   Returns the authenticated partner's current tier and progression info
// @Tags          Partner Auth
// @Accept        json
// @Produce       json
// @Success       200 {object} structs.ResponseBody{body=database.PartnerTier}
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/auth/tier [get]
func (co AuthController) TierDetails(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := co.Service.GetTierDetails(partner.ID)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

