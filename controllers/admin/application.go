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

// ApplicationController handles partner application review
type ApplicationController struct {
	common.Controller
	Service *adminsvc.ApplicationService
}

func (co ApplicationController) Register(router *gin.RouterGroup) {
	co.Service = &adminsvc.ApplicationService{
		BaseService: services.BaseService{DB: co.DB},
	}
	r := router.Use(middlewares.AdminAuth(co.DB), middlewares.HasPermission("manage_partners"))
	{
		r.POST("/list", co.List)
		r.GET("/detail/:id", co.Detail)
		r.POST("/:id/approve", co.Approve)
		r.POST("/:id/reject", co.Reject)
	}
}

// List
// @Summary       List partner applications
// @Description   Returns a paginated list of partner applications for admin review
// @Tags          Admin Applications
// @Accept        json
// @Produce       json
// @Param         request body structs.ApplicationListParams true "Filters and pagination"
// @Success       200 {object} structs.ResponseBody{body=structs.PaginationResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/applications/list [post]
func (co ApplicationController) List(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.ApplicationListParams
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
// @Summary       Get application detail
// @Description   Returns one partner application with full submission data
// @Tags          Admin Applications
// @Accept        json
// @Produce       json
// @Param         id path string true "Application ID"
// @Success       200 {object} structs.ResponseBody{body=database.PartnerApplication}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/applications/detail/{id} [get]
func (co ApplicationController) Detail(c *gin.Context) {
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

// Approve
// @Summary       Approve partner application
// @Description   Marks an application approved and provisions the partner account
// @Tags          Admin Applications
// @Accept        json
// @Produce       json
// @Param         id path string true "Application ID"
// @Success       200 {object} structs.ResponseBody{body=database.PartnerApplication}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/applications/{id}/approve [post]
func (co ApplicationController) Approve(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	admin := middlewares.AdminGetAuth(c)
	if admin == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := c.Param("id")
	if id == "" {
		co.SetError(c, http.StatusBadRequest, "id is required")
		return
	}

	result, err := co.Service.Approve(id, admin.ID)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// Reject
// @Summary       Reject partner application
// @Description   Marks an application rejected with a reason
// @Tags          Admin Applications
// @Accept        json
// @Produce       json
// @Param         id path string true "Application ID"
// @Param         request body structs.ApplicationReviewParams true "Rejection reason"
// @Success       200 {object} structs.ResponseBody{body=database.PartnerApplication}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/applications/{id}/reject [post]
func (co ApplicationController) Reject(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	admin := middlewares.AdminGetAuth(c)
	if admin == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := c.Param("id")
	if id == "" {
		co.SetError(c, http.StatusBadRequest, "id is required")
		return
	}

	var params structs.ApplicationReviewParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Reject(id, admin.ID, params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}
