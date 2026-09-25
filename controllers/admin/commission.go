package admin

import (
	"net/http"

	"xmeta-partner/controllers/common"
	internalCommission "xmeta-partner/internal/commission"
	"xmeta-partner/middlewares"
	"xmeta-partner/structs"

	"github.com/gin-gonic/gin"
)

type CommissionController struct {
	common.Controller
	Service *internalCommission.Service
}

func (co CommissionController) Register(router *gin.RouterGroup) {
	co.Service = internalCommission.NewService(co.DB)

	view := router.Use(middlewares.AdminAuth(co.DB), middlewares.HasPermission("view_partner_commissions"))
	{
		view.POST("/list", co.List)
	}

	manage := router.Use(middlewares.AdminAuth(co.DB), middlewares.HasPermission("manage_partner_commissions"))
	{
		manage.POST("/sync-futures-commissions", co.SyncFuturesCommissions)
	}
}

// List
// @Summary       List all commissions (admin)
// @Description   Returns a paginated list of all commissions across all partners
// @Tags          Admin Commissions
// @Accept        json
// @Produce       json
// @Param         request body structs.AdminCommissionListParams true "Filters and pagination"
// @Success       200 {object} structs.ResponseBody{body=structs.PaginationResponse}
// @Router        /admin/partner/commissions/list [post]
func (co CommissionController) List(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.AdminCommissionListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Queries.AdminListCommissions.Handle(params)
	if err != nil {
		co.SetError(c, http.StatusInternalServerError, err.Error())
		return
	}

	co.SetBody(c, result)
}

// SyncFuturesCommissions
// @Summary       Sync futures closed positions into partner commissions
// @Description   Reads futures_closed_positions directly and creates pending partner commissions for the matching referral window
// @Tags          Admin Commissions
// @Accept        json
// @Produce       json
// @Param         request body structs.FuturesClosedPositionSyncParams true "Closed position sync range"
// @Success       200 {object} structs.ResponseBody
// @Router        /admin/partner/commissions/sync-futures-commissions [post]
func (co CommissionController) SyncFuturesCommissions(c *gin.Context) {
	defer func() { c.JSON(co.GetBody(c)) }()

	var params structs.FuturesClosedPositionSyncParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := co.Service.Commands.SyncFuturesClosedPositions.Handle(params)
	if err != nil {
		co.SetError(c, http.StatusBadRequest, err.Error())
		return
	}

	co.SetBody(c, result)
}
