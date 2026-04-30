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

// ConfigController handles tier and commission config management
type ConfigController struct {
	common.Controller
	Service *adminsvc.ConfigService
}

func (co ConfigController) Register(router *gin.RouterGroup) {
	co.Service = &adminsvc.ConfigService{
		BaseService: services.BaseService{DB: co.DB},
	}
	r := router.Use(middlewares.AdminAuth(co.DB), middlewares.HasPermission("manage_partners"))
	{
		r.GET("/tiers", co.ListTiers)
		r.POST("/tiers", co.CreateTier)
		r.PUT("/tiers/:id", co.UpdateTier)
		r.DELETE("/tiers/:id", co.DeleteTier)
	}
}

// ListTiers
// @Summary       List commission tiers (admin)
// @Description   Returns all commission tiers for admin configuration
// @Tags          Admin Config
// @Accept        json
// @Produce       json
// @Success       200 {object} structs.ResponseBody{body=[]database.PartnerTier}
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/config/tiers [get]
func (co ConfigController) ListTiers(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	result, err := co.Service.ListTiers()
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// CreateTier
// @Summary       Create commission tier
// @Description   Creates a new commission tier configuration
// @Tags          Admin Config
// @Accept        json
// @Produce       json
// @Param         request body structs.TierCreateParams true "Tier create payload"
// @Success       200 {object} structs.ResponseBody{body=database.PartnerTier}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/config/tiers [post]
func (co ConfigController) CreateTier(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.TierCreateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.CreateTier(params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// UpdateTier
// @Summary       Update commission tier
// @Description   Updates an existing commission tier configuration
// @Tags          Admin Config
// @Accept        json
// @Produce       json
// @Param         id path string true "Tier ID"
// @Param         request body structs.TierUpdateParams true "Tier update payload"
// @Success       200 {object} structs.ResponseBody{body=database.PartnerTier}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/config/tiers/{id} [put]
func (co ConfigController) UpdateTier(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	id := c.Param("id")
	if id == "" {
		co.SetError(c, http.StatusBadRequest, "id is required")
		return
	}

	var params structs.TierUpdateParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.UpdateTier(id, params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// DeleteTier
// @Summary       Delete commission tier
// @Description   Removes a commission tier configuration
// @Tags          Admin Config
// @Accept        json
// @Produce       json
// @Param         id path string true "Tier ID"
// @Success       200 {object} structs.ResponseBody{body=structs.SuccessResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/config/tiers/{id} [delete]
func (co ConfigController) DeleteTier(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	id := c.Param("id")
	if id == "" {
		co.SetError(c, http.StatusBadRequest, "id is required")
		return
	}

	err := co.Service.DeleteTier(id)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, structs.SuccessResponse{Success: true})
}
