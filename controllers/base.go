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

// Register Controller
func Register(router *gin.RouterGroup) {
	database := db.Connect()

	// Activity log middleware
	router.Use(middlewares.ActivityLog(database))

	bc := common.Controller{DB: database}

	// Public routes (no auth)
	public.Controller{Controller: bc}.Register(router.Group("public/partner"))

	// Internal/system routes (API key auth — events from monorepo)
	system.EventsController{Controller: bc}.Register(router.Group("internal"))

	// Partner routes (partner Cognito auth)
	partnerGroup := router.Group("partner")
	partner.AuthController{Controller: bc}.Register(partnerGroup.Group("auth"))
	partner.DashboardController{Controller: bc}.Register(partnerGroup.Group("dashboard"))
	partner.ReferralController{Controller: bc}.Register(partnerGroup.Group("referrals"))
	partner.LinkController{Controller: bc}.Register(partnerGroup.Group("links"))
	partner.CommissionController{Controller: bc}.Register(partnerGroup.Group("commissions"))
	partner.PayoutController{Controller: bc}.Register(partnerGroup.Group("payouts"))
	partner.SubAffiliateController{Controller: bc}.Register(partnerGroup.Group("sub-affiliates"))

	// Admin routes (admin Cognito auth + RBAC)
	adminGroup := router.Group("admin/partner")
	admin.ApplicationController{Controller: bc}.Register(adminGroup.Group("applications"))
	admin.PartnerController{Controller: bc}.Register(adminGroup.Group("partners"))
	admin.ConfigController{Controller: bc}.Register(adminGroup.Group("config"))
	admin.PayoutController{Controller: bc}.Register(adminGroup.Group("payouts"))
	admin.AnalyticsController{Controller: bc}.Register(adminGroup.Group("analytics"))
}
