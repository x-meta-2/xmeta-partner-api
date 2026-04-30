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

// PayoutController handles admin payout management
type PayoutController struct {
	common.Controller
	Service *adminsvc.PayoutService
}

func (co PayoutController) Register(router *gin.RouterGroup) {
	co.Service = &adminsvc.PayoutService{
		BaseService: services.BaseService{DB: co.DB},
	}
	r := router.Use(middlewares.AdminAuth(co.DB), middlewares.HasPermission("manage_partners"))
	{
		r.POST("/list", co.List)
		r.GET("/detail/:id", co.Detail)
		r.POST("/pending", co.PendingList)
		r.POST("/:id/approve", co.Approve)
		r.POST("/:id/reject", co.Reject)
	}
}

// List
// @Summary       List payouts (admin)
// @Description   Returns a paginated list of payout records across all partners
// @Tags          Admin Payouts
// @Accept        json
// @Produce       json
// @Param         request body structs.PayoutListParams true "Filters and pagination"
// @Success       200 {object} structs.ResponseBody{body=structs.PaginationResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/payouts/list [post]
func (co PayoutController) List(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.PayoutListParams
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
// @Summary       Get payout detail (admin)
// @Description   Returns one payout with line items for admin review
// @Tags          Admin Payouts
// @Accept        json
// @Produce       json
// @Param         id path string true "Payout ID"
// @Success       200 {object} structs.ResponseBody{body=database.Payout}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/payouts/detail/{id} [get]
func (co PayoutController) Detail(c *gin.Context) {
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

// PendingList
// @Summary       List pending payouts
// @Description   Returns payouts awaiting admin approval, paginated
// @Tags          Admin Payouts
// @Accept        json
// @Produce       json
// @Param         request body structs.PayoutListParams true "Filters and pagination"
// @Success       200 {object} structs.ResponseBody{body=structs.PaginationResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/payouts/pending [post]
func (co PayoutController) PendingList(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.PayoutListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.PendingList(params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// Approve
// @Summary       Approve payout
// @Description   Approves a pending payout for processing
// @Tags          Admin Payouts
// @Accept        json
// @Produce       json
// @Param         id path string true "Payout ID"
// @Success       200 {object} structs.ResponseBody{body=database.Payout}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/payouts/{id}/approve [post]
func (co PayoutController) Approve(c *gin.Context) {
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
// @Summary       Reject payout
// @Description   Rejects a pending payout with a reason
// @Tags          Admin Payouts
// @Accept        json
// @Produce       json
// @Param         id path string true "Payout ID"
// @Param         request body structs.PayoutReviewParams true "Rejection reason"
// @Success       200 {object} structs.ResponseBody{body=database.Payout}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/payouts/{id}/reject [post]
func (co PayoutController) Reject(c *gin.Context) {
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

	var params structs.PayoutReviewParams
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
