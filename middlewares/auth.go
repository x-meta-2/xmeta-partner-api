package middlewares

import (
	"net/http"
	"strings"

	"xmeta-partner/database"
	"xmeta-partner/structs"
	"xmeta-partner/utils"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

const (
	AdminAuthKey         = "admin_auth"
	UserAuthKey          = "user_auth"
	PartnerAuthKey       = "partner_auth"
	InternalAPIKeyHeader = "X-Internal-API-Key"
)

const (
	ErrInvalidToken       = "INVALID_TOKEN"
	ErrUserNotFound       = "USER_NOT_FOUND"
	ErrNotAPartner        = "NOT_A_PARTNER"
	ErrApplicationPending = "APPLICATION_PENDING"
	ErrPartnerSuspended   = "PARTNER_SUSPENDED"
	ErrForbidden          = "FORBIDDEN"
)

type errorBody struct {
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Body    interface{} `json:"body"`
}

func abortWithCode(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, errorBody{Message: message, Code: code, Body: nil})
}

func extractBearerToken(c *gin.Context) string {
	header := c.Request.Header["Authorization"]
	if len(header) == 0 || len(header[0]) < 8 {
		return ""
	}
	token := header[0]
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}
	return strings.TrimSpace(token)
}

func decodeIdentity(token string, decode func(string) (map[string]interface{}, error)) (string, string, bool) {
	claims, err := decode(token)
	if err != nil {
		return "", "", false
	}
	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	if sub == "" {
		return "", "", false
	}
	return sub, email, true
}

func AdminAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		fail := structs.ResponseBody{Message: "Admin authentication failed"}

		token := extractBearerToken(c)
		if token == "" || utils.GetAdminCognitoService() == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, fail)
			return
		}

		userID, userEmail, ok := decodeIdentity(token, utils.DecodeAdminToken)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, fail)
			return
		}

		var adminUser database.AdminUser
		query := db.Preload("AdminGroup.Permissions")
		err := query.Where("id = ?", userID).First(&adminUser).Error
		if err != nil && userEmail != "" {
			err = query.Where("email = ?", userEmail).First(&adminUser).Error
		}
		if err != nil {
			adminUser = database.AdminUser{Base: database.Base{ID: userID}, Email: userEmail}
		}

		c.Set(AdminAuthKey, &adminUser)
		c.Next()
	}
}

func AdminGetAuth(c *gin.Context) *database.AdminUser {
	val, ok := c.Get(AdminAuthKey)
	if !ok {
		return nil
	}
	return val.(*database.AdminUser)
}

func UserAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" || utils.GetPartnerCognitoService() == nil {
			abortWithCode(c, http.StatusUnauthorized, ErrInvalidToken, "Missing or invalid token")
			return
		}

		userID, _, ok := decodeIdentity(token, utils.DecodePartnerToken)
		if !ok {
			abortWithCode(c, http.StatusUnauthorized, ErrInvalidToken, "Invalid token")
			return
		}

		var user database.User
		if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
			abortWithCode(c, http.StatusUnauthorized, ErrUserNotFound, "User not found")
			return
		}

		c.Set(UserAuthKey, &user)
		c.Next()
	}
}

func UserGetAuth(c *gin.Context) *database.User {
	val, ok := c.Get(UserAuthKey)
	if !ok {
		return nil
	}
	return val.(*database.User)
}

func PartnerAuth(db *gorm.DB) gin.HandlerFunc {
	demoMode := viper.GetBool("PARTNER_DEMO_MODE")

	return func(c *gin.Context) {
		if utils.GetPartnerCognitoService() == nil {
			if !demoMode {
				abortWithCode(c, http.StatusUnauthorized, ErrInvalidToken, "Authentication service unavailable")
				return
			}
			token := extractBearerToken(c)
			if token == "" {
				abortWithCode(c, http.StatusUnauthorized, ErrInvalidToken, "Missing token")
				return
			}
			var partner database.Partner
			if err := db.Preload("Tier").Where("status = ?", database.PartnerStatusActive).
				Order("created_at ASC").First(&partner).Error; err != nil {
				abortWithCode(c, http.StatusUnauthorized, ErrInvalidToken, "No active partner for demo")
				return
			}
			c.Set(PartnerAuthKey, &partner)
			c.Next()
			return
		}

		token := extractBearerToken(c)
		if token == "" {
			abortWithCode(c, http.StatusUnauthorized, ErrInvalidToken, "Missing token")
			return
		}
		userID, _, ok := decodeIdentity(token, utils.DecodePartnerToken)
		if !ok {
			abortWithCode(c, http.StatusUnauthorized, ErrInvalidToken, "Invalid token")
			return
		}

		var partner database.Partner
		err := db.Preload("User").Preload("Tier").
			Where("user_id = ?", userID).First(&partner).Error
		if err != nil {
			var app database.PartnerApplication
			appErr := db.Where("user_id = ?", userID).
				Order("created_at DESC").First(&app).Error
			if appErr == nil && app.Status == "pending" {
				abortWithCode(c, http.StatusForbidden, ErrApplicationPending, "Your partner application is under review")
				return
			}
			abortWithCode(c, http.StatusForbidden, ErrNotAPartner, "You are not a partner yet")
			return
		}
		if partner.Status != database.PartnerStatusActive {
			abortWithCode(c, http.StatusForbidden, ErrPartnerSuspended, "Partner account is not active")
			return
		}

		c.Set(PartnerAuthKey, &partner)
		c.Next()
	}
}

func PartnerGetAuth(c *gin.Context) *database.Partner {
	val, ok := c.Get(PartnerAuthKey)
	if !ok {
		return nil
	}
	return val.(*database.Partner)
}

func InternalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader(InternalAPIKeyHeader)
		expectedKey := viper.GetString("INTERNAL_API_KEY")

		if apiKey == "" || expectedKey == "" || apiKey != expectedKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ResponseBody{Message: "Invalid internal API key"})
			return
		}
		c.Next()
	}
}
