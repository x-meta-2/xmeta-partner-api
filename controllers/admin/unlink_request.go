package admin

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

type UnlinkRequestController struct {
	common.Controller
	Service *internalReferral.Service
}

func (co UnlinkRequestController) Register(router *gin.RouterGroup) {
	co.Service = internalReferral.NewService(co.DB)
	r := router.Use(middlewares.AdminAuth(co.DB), middlewares.HasPermission("manage_partners"))
	{
		r.POST("/list", co.List)
		r.POST("/:id/approve", co.Approve)
		r.POST("/:id/reject", co.Reject)
	}
}

// List
// @Summary       List referral unlink requests
// @Description   Returns paginated user requests to stop following their current partner.
// @Tags          Admin Referral Unlink Requests
// @Accept        json
// @Produce       json
// @Param         request body structs.ReferralUnlinkRequestListParams true "Filters and pagination"
// @Success       200 {object} structs.ResponseBody{body=structs.PaginationResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/referral-unlink-requests/list [post]
func (co UnlinkRequestController) List(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.ReferralUnlinkRequestListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Queries.AdminListUnlinkRequests.Handle(params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// Approve
// @Summary       Approve referral unlink request
// @Description   Approves a pending request and unlinks the user from their active partner.
// @Tags          Admin Referral Unlink Requests
// @Accept        json
// @Produce       json
// @Param         id path string true "Request ID"
// @Param         request body structs.ReferralUnlinkRequestReviewParams true "Review note"
// @Success       200 {object} structs.ResponseBody{body=database.ReferralUnlinkRequest}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/referral-unlink-requests/{id}/approve [post]
func (co UnlinkRequestController) Approve(c *gin.Context) {
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

	var params structs.ReferralUnlinkRequestReviewParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Commands.ApproveUnlinkRequest.Handle(id, admin.ID, params)
	if err != nil {
		co.setReviewError(c, err)
		return
	}

	co.SetBody(c, result)
}

// Reject
// @Summary       Reject referral unlink request
// @Description   Rejects a pending request and keeps the user's partner relationship active.
// @Tags          Admin Referral Unlink Requests
// @Accept        json
// @Produce       json
// @Param         id path string true "Request ID"
// @Param         request body structs.ReferralUnlinkRequestReviewParams true "Review note"
// @Success       200 {object} structs.ResponseBody{body=database.ReferralUnlinkRequest}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /admin/partner/referral-unlink-requests/{id}/reject [post]
func (co UnlinkRequestController) Reject(c *gin.Context) {
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

	var params structs.ReferralUnlinkRequestReviewParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Commands.RejectUnlinkRequest.Handle(id, admin.ID, params)
	if err != nil {
		co.setReviewError(c, err)
		return
	}

	co.SetBody(c, result)
}

func (co UnlinkRequestController) setReviewError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrUnlinkRequestNotFound):
		co.SetError(c, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrUnlinkRequestReviewed),
		errors.Is(err, domain.ErrNoActiveReferral):
		co.SetError(c, http.StatusBadRequest, err.Error())
	default:
		co.SetError(c, http.StatusInternalServerError, err.Error())
	}
}
