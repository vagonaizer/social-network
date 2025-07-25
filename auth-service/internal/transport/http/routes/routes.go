package routes

import (
	"github.com/gin-gonic/gin"
	files "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "social-network/auth-service/docs" // Импорт docs
	"social-network/auth-service/internal/transport/http/handlers"
	"social-network/auth-service/internal/transport/http/middleware"
)

// @title Auth Service API
// @version 1.0
// @description Authentication and Authorization Service API documentation
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func SetupRoutes(
	router *gin.Engine,
	authHandler *handlers.AuthHandler,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Debug endpoint
	router.GET("/debug", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Server is running",
			"swagger": "Available at /swagger/index.html",
			"routes": []string{
				"GET /swagger/index.html",
				"GET /health",
				"POST /api/auth/register",
				"POST /api/auth/login",
			},
		})
	})

	// Swagger documentation - используем правильный импорт
	router.GET("/swagger/*any", ginSwagger.WrapHandler(files.Handler))

	// Redirect root to swagger
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "alive"})
	})

	router.GET("/health/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ready"})
	})

	// API routes
	api := router.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			// Public endpoints
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/verify-email", authHandler.VerifyEmail)
			auth.POST("/reset-password", authHandler.InitiatePasswordReset)
			auth.POST("/reset-password/confirm", authHandler.ResetPassword)

			// Protected endpoints
			protected := auth.Group("")
			protected.Use(authMiddleware.RequireAuth())
			{
				protected.GET("/me", authHandler.GetCurrentUser)
				protected.PUT("/change-password", authHandler.ChangePassword)
				protected.POST("/logout", authHandler.Logout)
				protected.GET("/validate", authHandler.ValidateToken)
			}

			// Admin endpoints
			admin := auth.Group("/users")
			admin.Use(authMiddleware.RequireAuth(), authMiddleware.RequireAdmin())
			{
				admin.POST("/:user_id/roles", authHandler.AssignRole)
				admin.DELETE("/:user_id/roles/:role", authHandler.RevokeRole)
				admin.GET("/:user_id/roles", authHandler.GetUserRoles)
			}
		}
	}
}
