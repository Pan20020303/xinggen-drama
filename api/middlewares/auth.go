package middlewares

import (
	"strings"

	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey   = "user_id"
	ContextUserRoleKey = "user_role"
)

func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := parseBearerClaims(c, authService)
		if !ok {
			return
		}

		if !hasAudience(claims.Audience, "user") {
			response.Forbidden(c, "user audience required")
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUserRoleKey, string(claims.Role))
		c.Next()
	}
}

func AdminAuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := parseBearerClaims(c, authService)
		if !ok {
			return
		}

		if !hasAudience(claims.Audience, "admin") {
			response.Forbidden(c, "admin audience required")
			c.Abort()
			return
		}
		if !services.IsPlatformAdminRole(claims.Role) {
			response.Forbidden(c, "admin role required")
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextUserRoleKey, string(claims.Role))
		c.Next()
	}
}

func parseBearerClaims(c *gin.Context, authService *services.AuthService) (*services.TokenClaims, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response.Unauthorized(c, "missing authorization header")
		c.Abort()
		return nil, false
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		response.Unauthorized(c, "invalid authorization header")
		c.Abort()
		return nil, false
	}

	claims, err := authService.ParseToken(parts[1])
	if err != nil {
		response.Unauthorized(c, "invalid or expired token")
		c.Abort()
		return nil, false
	}
	return claims, true
}

func hasAudience(audiences []string, target string) bool {
	for _, aud := range audiences {
		if aud == target {
			return true
		}
	}
	return false
}

func GetUserID(c *gin.Context) (uint, bool) {
	v, ok := c.Get(ContextUserIDKey)
	if !ok {
		return 0, false
	}
	uid, ok := v.(uint)
	return uid, ok
}
