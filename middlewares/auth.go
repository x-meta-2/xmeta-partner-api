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

// Error codes returned by auth middlewares — consumed by the frontend to
// decide which screen to show (login / apply / pending / suspended / dashboard).
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

// extractBearerToken — Authorization header-аас Bearer token гаргаж авна
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

// decodeIdentity — Cognito token-оос sub, email-ийг задлан гаргана
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

// AdminAuth — admin Cognito pool-оос validate хийгээд admin_users table-аас admin-ыг олно
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

// AdminGetAuth — context-оос authenticated admin-ыг авна
func AdminGetAuth(c *gin.Context) *database.AdminUser {
	val, ok := c.Get(AdminAuthKey)
	if !ok {
		return nil
	}
	return val.(*database.AdminUser)
}

// UserAuth — Cognito token-ыг validate хийж, xmeta `users` table-аас user-ыг
// ачаална. Partner эрхгүй байсан ч нэвтрэх боломж өгнө (apply/status endpoint-д
// зориулагдсан).
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

// UserGetAuth — context-оос authenticated xmeta user-ыг авна
func UserGetAuth(c *gin.Context) *database.User {
	val, ok := c.Get(UserAuthKey)
	if !ok {
		return nil
	}
	return val.(*database.User)
}

// PartnerAuth — UserAuth + зайлшгүй active partner шаардана. Partner row эсвэл
// status-аас хамааран тусгай code-тойгоор 403 буцаана → frontend зохих
// onboarding page-руу чиглүүлнэ.
func PartnerAuth(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Demo mode — Cognito тохируулаагүй үед эхний идэвхтэй partner-ыг хэрэглэнэ
		if utils.GetPartnerCognitoService() == nil {
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
			// No partner row → determine if application exists
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

// PartnerGetAuth — context-оос authenticated partner-ыг авна
func PartnerGetAuth(c *gin.Context) *database.Partner {
	val, ok := c.Get(PartnerAuthKey)
	if !ok {
		return nil
	}
	return val.(*database.Partner)
}

// InternalAuth — service-to-service дуудалтыг X-Internal-API-Key header-аар шалгана
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
