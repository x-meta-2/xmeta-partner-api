package partner

import (
	"net/http"

	"xmeta-partner/controllers/common"
	internalPayout "xmeta-partner/internal/payout"
	"xmeta-partner/middlewares"
	"xmeta-partner/structs"

	"github.com/gin-gonic/gin"
)

type PayoutController struct {
	common.Controller
	Service *internalPayout.Service
}

func (co PayoutController) Register(router *gin.RouterGroup) {
	co.Service = internalPayout.NewService(co.DB)
	r := router.Use(middlewares.PartnerAuth(co.DB))
	{
		r.POST("/list", co.List)
		r.GET("/detail/:id", co.Detail)
		r.GET("/pending", co.Pending)
		r.POST("/request", co.Request)
	}
}

// List
// @Summary       List partner payouts
// @Description   Returns a paginated list of payout records for the partner
// @Tags          Partner Payouts
// @Accept        json
// @Produce       json
// @Param         request body structs.PayoutListParams true "Filters and pagination"
// @Success       200 {object} structs.ResponseBody{body=structs.PaginationResponse}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/payouts/list [post]
func (co PayoutController) List(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var params structs.PayoutListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Queries.ListPayouts.Handle(partner.ID, params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// Detail
// @Summary       Get payout detail
// @Description   Returns one payout record with its line items
// @Tags          Partner Payouts
// @Accept        json
// @Produce       json
// @Param         id path string true "Payout ID"
// @Success       200 {object} structs.ResponseBody{body=database.Payout}
// @Failure       400 {object} structs.ErrorResponse
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/payouts/detail/{id} [get]
func (co PayoutController) Detail(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	id := c.Param("id")
	if id == "" {
		co.SetError(c, http.StatusBadRequest, "id is required")
		return
	}

	result, err := co.Service.Queries.PayoutDetail.Handle(partner.ID, id)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// Pending
// @Summary       Pending payout balance
// @Description   Returns the partner's currently pending payout balance and breakdown
// @Tags          Partner Payouts
// @Accept        json
// @Produce       json
// @Success       200 {object} structs.ResponseBody
// @Failure       401 {object} structs.ErrorResponse
// @Failure       500 {object} structs.ErrorResponse
// @Security      BearerAuth
// @Router        /partner/payouts/pending [get]
func (co PayoutController) Pending(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := co.Service.Queries.PendingCommissions.Handle(partner.ID)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

func (co PayoutController) Request(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	partner := middlewares.PartnerGetAuth(c)
	if partner == nil {
		co.SetError(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	payout, err := co.Service.Commands.RequestPayout.Handle(partner.ID)
	if err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	co.SetBody(c, payout)
}
