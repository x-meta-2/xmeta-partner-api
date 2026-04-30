package middlewares

import (
	"net/http"

	"xmeta-partner/structs"

	"github.com/gin-gonic/gin"
)

const SuperAdminGroupID = "4"

// HasPermission checks if admin user has the required permission
func HasPermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminUser := AdminGetAuth(c)
		if adminUser == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, structs.ResponseBody{
				Message: "Admin authentication required",
			})
			return
		}

		// Super Admin bypass
		if adminUser.AdminGroupID == SuperAdminGroupID {
			c.Next()
			return
		}

		if adminUser.AdminGroup == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, structs.ResponseBody{
				Message: "Permission denied (no group assigned)",
			})
			return
		}

		has := false
		for _, p := range adminUser.AdminGroup.Permissions {
			if p.Name == permission || p.ID == permission {
				has = true
				break
			}
		}

		if !has {
			c.AbortWithStatusJSON(http.StatusForbidden, structs.ResponseBody{
				Message: "Permission denied: " + permission,
			})
			return
		}

		c.Next()
	}
}
