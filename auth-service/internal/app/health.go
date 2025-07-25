package app

import (
	"context"
	"fmt"
	"time"
)

// HealthChecker интерфейс для проверки здоровья компонентов
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// DatabaseHealthChecker проверяет состояние базы данных
type DatabaseHealthChecker struct {
	app *App
}

func NewDatabaseHealthChecker(app *App) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{app: app}
}

func (h *DatabaseHealthChecker) HealthCheck(ctx context.Context) error {
	if h.app.database == nil {
		return fmt.Errorf("database connection not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return h.app.database.Ping(ctx)
}

// DetailedHealth возвращает детальную информацию о состоянии приложения
func (a *App) DetailedHealth() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"service":   a.config.Logger.ServiceName,
		"uptime":    time.Since(time.Now()).String(),
	}

	// Детальная проверка компонентов
	components := make(map[string]interface{})

	// Проверка базы данных
	dbChecker := NewDatabaseHealthChecker(a)
	if err := dbChecker.HealthCheck(ctx); err != nil {
		components["database"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	} else {
		// Дополнительная информация о БД
		stats := a.database.Stat()
		components["database"] = map[string]interface{}{
			"status":               "healthy",
			"total_connections":    stats.TotalConns(),
			"idle_connections":     stats.IdleConns(),
			"acquired_connections": stats.AcquiredConns(),
		}
	}

	// Проверка конфигурации
	components["config"] = map[string]interface{}{
		"status":    "healthy",
		"http_port": a.config.Server.HTTP.Port,
		"grpc_port": a.config.Server.GRPC.Port,
		"log_level": a.config.Logger.Level,
	}

	health["components"] = components

	// Определяем общий статус
	overallHealthy := true
	for _, component := range components {
		if comp, ok := component.(map[string]interface{}); ok {
			if status, exists := comp["status"]; exists && status != "healthy" {
				overallHealthy = false
				break
			}
		}
	}

	if !overallHealthy {
		health["status"] = "unhealthy"
	}

	return health
}
