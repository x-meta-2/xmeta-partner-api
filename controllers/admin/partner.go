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

// PartnerController handles admin partner management
type PartnerController struct {
	common.Controller
	Service *adminsvc.PartnerService
}

func (co PartnerController) Register(router *gin.RouterGroup) {
	co.Service = &adminsvc.PartnerService{
		BaseService: services.BaseService{DB: co.DB},
	}
	r := router.Use(middlewares.AdminAuth(co.DB), middlewares.HasPermission("manage_partners"))
	{
		r.POST("/list", co.List)
		r.GET("/detail/:id", co.Detail)
		r.PUT("/:id/tier", co.UpdateTier)
		r.PUT("/:id/status", co.UpdateStatus)
	}
}

// List
// @Summary       List partners
// @Description   Returns a paginated list of partners for admin management
// @Tags          Admin Partners
// @Accept        json
// @Produce       json
// @Param         request body structs.PartnerListParams true "Filters and pagination"
// @Success       200 {object} structs.ResponseBody{body=structs.PaginationResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/partners/list [post]
func (co PartnerController) List(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.PartnerListParams
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
// @Summary       Get partner detail
// @Description   Returns one partner with related stats and tier info
// @Tags          Admin Partners
// @Accept        json
// @Produce       json
// @Param         id path string true "Partner ID"
// @Success       200 {object} structs.ResponseBody{body=database.Partner}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/partners/detail/{id} [get]
func (co PartnerController) Detail(c *gin.Context) {
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

// UpdateTier
// @Summary       Update partner tier
// @Description   Assigns the partner to a different commission tier
// @Tags          Admin Partners
// @Accept        json
// @Produce       json
// @Param         id path string true "Partner ID"
// @Param         request body object true "Tier update payload (tierId)"
// @Success       200 {object} structs.ResponseBody{body=database.Partner}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/partners/{id}/tier [put]
func (co PartnerController) UpdateTier(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	id := c.Param("id")
	if id == "" {
		co.SetError(c, http.StatusBadRequest, "id is required")
		return
	}

	var params struct {
		TierID string `json:"tierId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.UpdateTier(id, params.TierID)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// UpdateStatus
// @Summary       Update partner status
// @Description   Activates, suspends, or otherwise changes the partner status
// @Tags          Admin Partners
// @Accept        json
// @Produce       json
// @Param         id path string true "Partner ID"
// @Param         request body object true "Status update payload (status)"
// @Success       200 {object} structs.ResponseBody{body=database.Partner}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/partners/{id}/status [put]
func (co PartnerController) UpdateStatus(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	id := c.Param("id")
	if id == "" {
		co.SetError(c, http.StatusBadRequest, "id is required")
		return
	}

	var params struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.UpdateStatus(id, params.Status)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}
