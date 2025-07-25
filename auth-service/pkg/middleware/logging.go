package middleware

import (
	"fmt"
	"time"

	"social-network/auth-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

// LoggingMiddleware создает middleware для логирования HTTP запросов
func LoggingMiddleware(zapLogger *logger.ZapLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Обрабатываем запрос
		c.Next()

		// Логируем после обработки
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		zapLogger.Info("HTTP Request",
			logger.String("method", method),
			logger.String("path", path),
			logger.Int("status", statusCode),
			logger.String("client_ip", clientIP),
			logger.Duration("latency", latency),
			logger.Int("body_size", bodySize),
			logger.String("user_agent", c.Request.UserAgent()),
		)

		// Логируем ошибки отдельно
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				zapLogger.Error("Request Error",
					logger.String("method", method),
					logger.String("path", path),
					logger.Error(err.Err),
					logger.String("type", fmt.Sprintf("%v", err.Type)),
				)
			}
		}
	}
}

// RecoveryMiddleware создает middleware для обработки паник
func RecoveryMiddleware(zapLogger *logger.ZapLogger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		zapLogger.Error("Panic recovered",
			logger.String("method", c.Request.Method),
			logger.String("path", c.Request.URL.Path),
			logger.String("client_ip", c.ClientIP()),
			logger.Any("panic", recovered),
			logger.String("stack", logger.GetCaller(3)),
		)

		c.AbortWithStatus(500)
	})
}
