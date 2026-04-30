package middlewares

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// CORS returns a gin.HandlerFunc that handles CORS using gin-contrib/cors
func CORS(c *gin.Context) {
	allowedOrigins := viper.GetString("ALLOWED_ORIGINS")

	config := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "Accept", "X-CSRF-Token", "X-Requested-With", "x-sync-source", "Cache-Control", "X-Internal-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	if allowedOrigins == "" || allowedOrigins == "*" {
		config.AllowAllOrigins = true
	} else {
		// Split and normalize origins (remove trailing slashes to be robust)
		rawOrigins := strings.Split(allowedOrigins, ",")
		for _, o := range rawOrigins {
			normalized := strings.TrimRight(strings.TrimSpace(o), "/")
			if normalized != "" {
				config.AllowOrigins = append(config.AllowOrigins, normalized)
				// Add the version with trailing slash too just in case
				config.AllowOrigins = append(config.AllowOrigins, normalized+"/")
			}
		}
	}

	// Create and execute the middleware
	mw := cors.New(config)
	mw(c)
}
