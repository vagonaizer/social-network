package middleware

import (
	"net/http"
	"social-network/auth-service/internal/service"
	"social-network/auth-service/internal/transport/http/dto"
	"strings"

	"github.com/gin-gonic/gin"

	"time"
)

type AuthMiddleware struct {
	jwtService *service.JWTService
}

func NewAuthMiddleware(jwtService *service.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// RequireAuth проверяет наличие и валидность access token
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			m.respondUnauthorized(c, "Missing authorization token")
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			m.respondUnauthorized(c, "Invalid or expired token")
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_username", claims.Username)
		c.Set("user_roles", claims.Roles)
		c.Set("user_verified", claims.IsVerified)

		c.Next()
	}
}

// RequireRole проверяет наличие определенной роли
func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("user_roles")
		if !exists {
			m.respondForbidden(c, "No roles found")
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			m.respondForbidden(c, "Invalid roles format")
			return
		}

		hasRole := false
		for _, role := range userRoles {
			if role == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			m.respondForbidden(c, "Insufficient permissions")
			return
		}

		c.Next()
	}
}

// RequireAdmin проверяет права администратора
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return m.RequireRole("admin")
}

// RequireModerator проверяет права модератора или выше
func (m *AuthMiddleware) RequireModerator() gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := c.Get("user_roles")
		if !exists {
			m.respondForbidden(c, "No roles found")
			return
		}

		userRoles, ok := roles.([]string)
		if !ok {
			m.respondForbidden(c, "Invalid roles format")
			return
		}

		hasPermission := false
		for _, role := range userRoles {
			if role == "admin" || role == "moderator" {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			m.respondForbidden(c, "Insufficient permissions")
			return
		}

		c.Next()
	}
}

// RequireVerified проверяет, что email пользователя подтвержден
func (m *AuthMiddleware) RequireVerified() gin.HandlerFunc {
	return func(c *gin.Context) {
		verified, exists := c.Get("user_verified")
		if !exists {
			m.respondForbidden(c, "Verification status unknown")
			return
		}

		isVerified, ok := verified.(bool)
		if !ok || !isVerified {
			m.respondForbidden(c, "Email verification required")
			return
		}

		c.Next()
	}
}

// OptionalAuth проверяет токен, но не требует его наличия
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			c.Next()
			return
		}

		// Сохраняем данные пользователя в контексте
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_username", claims.Username)
		c.Set("user_roles", claims.Roles)
		c.Set("user_verified", claims.IsVerified)

		c.Next()
	}
}

func (m *AuthMiddleware) extractToken(c *gin.Context) string {
	// Проверяем заголовок Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Проверяем query параметр
	if token := c.Query("token"); token != "" {
		return token
	}

	return ""
}

func (m *AuthMiddleware) respondUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
		Error:     "unauthorized",
		Message:   message,
		Timestamp: time.Now(),
		Path:      c.Request.URL.Path,
	})
	c.Abort()
}

func (m *AuthMiddleware) respondForbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, dto.ErrorResponse{
		Error:     "forbidden",
		Message:   message,
		Timestamp: time.Now(),
		Path:      c.Request.URL.Path,
	})
	c.Abort()
}
