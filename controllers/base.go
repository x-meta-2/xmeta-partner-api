package controllers

import (
	"xmeta-partner/controllers/admin"
	"xmeta-partner/controllers/common"
	"xmeta-partner/controllers/partner"
	"xmeta-partner/controllers/public"
	"xmeta-partner/controllers/system"
	db "xmeta-partner/database"
	"xmeta-partner/middlewares"

	"github.com/gin-gonic/gin"
)

func Register(router *gin.RouterGroup) {
	database := db.Connect()
	router.Use(middlewares.ActivityLog(database))

	bc := common.Controller{DB: database}

	public.Controller{Controller: bc}.Register(router.Group("public/partner"))
	system.EventsController{Controller: bc}.Register(router.Group("internal"))

	partnerGroup := router.Group("partner")
	partner.AuthController{Controller: bc}.Register(partnerGroup.Group("auth"))
	partner.DashboardController{Controller: bc}.Register(partnerGroup.Group("dashboard"))
	partner.ReferralController{Controller: bc}.Register(partnerGroup.Group("referrals"))
	partner.LinkController{Controller: bc}.Register(partnerGroup.Group("links"))
	partner.CommissionController{Controller: bc}.Register(partnerGroup.Group("commissions"))
	partner.PayoutController{Controller: bc}.Register(partnerGroup.Group("payouts"))

	adminGroup := router.Group("admin/partner")
	admin.ApplicationController{Controller: bc}.Register(adminGroup.Group("applications"))
	admin.PartnerController{Controller: bc}.Register(adminGroup.Group("partners"))
	admin.ReferralController{Controller: bc}.Register(adminGroup.Group("referrals"))
	admin.ConfigController{Controller: bc}.Register(adminGroup.Group("config"))
	admin.PayoutController{Controller: bc}.Register(adminGroup.Group("payouts"))
	admin.AnalyticsController{Controller: bc}.Register(adminGroup.Group("analytics"))
	admin.CommissionController{Controller: bc}.Register(adminGroup.Group("commissions"))
}
