package main

import (
	"fmt"
	"log"
	"net/http"

	"xmeta-partner/controllers"
	_ "xmeta-partner/docs" // swag-аар generate хийгдэх OpenAPI документ
	"xmeta-partner/middlewares"
	"xmeta-partner/utils"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           xmeta Partner API
// @version         1.0
// @description     Partner program backend — public, partner, admin, system endpoints.
// @description     Authentication is split into three layers:
// @description       • BearerAuth (PartnerAuth/AdminAuth): Cognito ID token in `Authorization: Bearer …`.
// @description       • InternalKey: `X-Internal-Key` header — only the xmeta-monorepo server should send this.
// @host            localhost:8090
// @BasePath        /api/v1
// @schemes         http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description "Bearer {Cognito ID Token}" format

// @securityDefinitions.apikey InternalKey
// @in header
// @name X-Internal-Key
// @description Server-to-server API key shared with xmeta-monorepo

func main() {
	utils.LoadConfig()

	if err := utils.InitCognito(); err != nil {
		log.Printf("Warning: Failed to initialize Cognito: %v", err)
	}

	gin.SetMode(gin.DebugMode)
	app := gin.Default()
	app.Use(middlewares.CORS)
	app.Use(gzip.Gzip(gzip.DefaultCompression))

	app.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"alive": true, "ready": true})
	})

	// Swagger UI: http://localhost:8090/swagger/index.html
	app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	controllers.Register(app.Group("/api/v1"))

	log.Fatalln(app.Run(fmt.Sprintf("0.0.0.0:%d", viper.GetInt("APP_PORT"))))
}
