package middleware

import (
	"strings"
	"stationery-management/internal/config"
	"stationery-management/pkg/jwt"
	"stationery-management/pkg/response"

	"github.com/gin-gonic/gin"
)

func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization token required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "Invalid authorization format")
			c.Abort()
			return
		}

		claims, err := jwt.ValidateToken(parts[1], cfg.JWTSecret)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Save claims in gin context
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)
		c.Set("userBranchID", claims.BranchID)
		c.Set("approverAccessType", claims.ApproverAccessType)
		c.Set("firstLogin", claims.FirstLogin)

		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			response.Forbidden(c, "Access denied")
			c.Abort()
			return
		}

		roleStr := userRole.(string)
		allowed := false
		for _, r := range roles {
			if r == roleStr {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Forbidden(c, "You do not have permission to perform this action")
			c.Abort()
			return
		}

		c.Next()
	}
}
