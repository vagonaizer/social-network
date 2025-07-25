package http

import (
	"context"
	"fmt"
	"net/http"
	"social-network/auth-service/internal/config"
	"social-network/auth-service/internal/service"
	"social-network/auth-service/internal/transport/http/handlers"
	httpMiddleware "social-network/auth-service/internal/transport/http/middleware"
	"social-network/auth-service/internal/transport/http/routes"
	"social-network/auth-service/pkg/logger"
	"social-network/auth-service/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type Server struct {
	server *http.Server
	logger logger.Logger
	config *config.Config
}

func NewServer(
	cfg *config.Config,
	authService *service.AuthService,
	jwtService *service.JWTService,
	validationService *service.ValidationService,
	customLogger logger.Logger,
	zapLogger *logger.ZapLogger,
) *Server {
	// Настройка Gin
	if cfg.Logger.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// Middleware
	router.Use(middleware.LoggingMiddleware(zapLogger))
	router.Use(middleware.RecoveryMiddleware(zapLogger))
	router.Use(gin.Recovery())

	// CORS middleware - ИСПРАВЛЕННАЯ ВЕРСИЯ
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Разрешаем запросы с фронтенда
		if origin == "http://localhost:3000" || origin == "http://127.0.0.1:3000" {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Handlers
	authHandler := handlers.NewAuthHandler(authService, jwtService, validationService, customLogger)
	authMiddleware := httpMiddleware.NewAuthMiddleware(jwtService)

	// Routes
	routes.SetupRoutes(router, authHandler, authMiddleware)

	// HTTP Server
	server := &http.Server{
		Addr:         ":" + cfg.Server.HTTP.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.HTTP.ReadTimeout,
		WriteTimeout: cfg.Server.HTTP.WriteTimeout,
		IdleTimeout:  cfg.Server.HTTP.IdleTimeout,
	}

	return &Server{
		server: server,
		logger: customLogger,
		config: cfg,
	}
}

func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server",
		logger.String("port", s.config.Server.HTTP.Port),
		logger.String("address", s.server.Addr),
	)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")

	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to stop HTTP server: %w", err)
	}

	s.logger.Info("HTTP server stopped")
	return nil
}
