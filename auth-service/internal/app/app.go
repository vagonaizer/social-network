package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"social-network/auth-service/internal/config"
	database "social-network/auth-service/internal/infrastructure/db"
	"social-network/auth-service/internal/service"
	"social-network/auth-service/pkg/logger"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	grpcTransport "social-network/auth-service/internal/transport/grpc"
	httpTransport "social-network/auth-service/internal/transport/http"
)

// App представляет основное приложение
type App struct {
	config     *config.Config
	logger     logger.Logger
	zapLogger  *logger.ZapLogger
	httpServer *httpTransport.Server
	grpcServer *grpcTransport.Server
	database   *database.Database

	// Сервисы
	authService       *service.AuthService
	jwtService        *service.JWTService
	validationService *service.ValidationService

	// Контекст для graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// New создает новый экземпляр приложения
func New() *App {
	ctx, cancel := context.WithCancel(context.Background())

	return &App{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Initialize инициализирует все компоненты приложения
func (a *App) Initialize() error {
	// 1. Загружаем конфигурацию
	if err := a.loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 2. Инициализируем логгеры
	if err := a.initLoggers(); err != nil {
		return fmt.Errorf("failed to initialize loggers: %w", err)
	}

	a.logger.Info("Starting application initialization",
		logger.String("service", a.config.Logger.ServiceName),
		logger.String("version", "1.0.0"),
	)

	// 3. Инициализируем базу данных
	if err := a.initDatabase(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	fmt.Printf("Connecting to DB on port: %s\n", os.Getenv("DB_PORT"))

	// 4. Инициализируем сервисы
	if err := a.initServices(); err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	// 5. Инициализируем транспортные слои
	if err := a.initTransports(); err != nil {
		return fmt.Errorf("failed to initialize transports: %w", err)
	}

	a.logger.Info("Application initialization completed successfully")
	return nil
}

// Run запускает приложение
func (a *App) Run() error {
	a.logger.Info("Starting application",
		logger.String("http_port", a.config.Server.HTTP.Port),
		logger.String("grpc_port", a.config.Server.GRPC.Port),
	)

	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Запускаем HTTP сервер
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.httpServer.Start(); err != nil {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Запускаем gRPC сервер
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.grpcServer.Start(); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	a.logger.Info("All servers started successfully")

	// Ожидаем сигнал завершения или ошибку
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		a.logger.Info("Received shutdown signal", logger.String("signal", sig.String()))
	case err := <-errChan:
		a.logger.Error("Server error occurred", logger.Error(err))
		return err
	case <-a.ctx.Done():
		a.logger.Info("Application context cancelled")
	}

	// Graceful shutdown
	return a.shutdown()
}

// Shutdown выполняет graceful shutdown приложения
func (a *App) Shutdown() error {
	a.cancel()
	return a.shutdown()
}

// Приватные методы инициализации

func (a *App) loadConfig() error {
	a.config = config.Load()
	return nil
}

func (a *App) initLoggers() error {
	a.logger = logger.NewCustomLogger(
		a.config.Logger.ServiceName,
		a.config.Logger.Level,
		nil,
	)

	a.zapLogger = logger.NewZapLogger(
		a.config.Logger.ServiceName,
		a.config.Logger.Level,
	)

	return nil
}

func (a *App) initDatabase() error {
	// Инициализируем базу данных
	db, err := database.NewDatabase(&a.config.Database, a.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	a.database = db

	// Выполняем health check
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.HealthCheck(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	a.logger.Info("Database initialized and health check passed",
		logger.String("host", a.config.Database.Host),
		logger.String("database", a.config.Database.DBName),
		logger.Int("max_connections", a.config.Database.MaxConnections),
	)

	return nil
}

func (a *App) initServices() error {
	// JWT сервис
	a.jwtService = service.NewJWTService(
		[]byte(a.config.JWT.AccessSecret),
		[]byte(a.config.JWT.RefreshSecret),
		a.config.JWT.Issuer,
	)

	// Сервис валидации
	a.validationService = service.NewValidationService()

	// Сервис аутентификации с использованием builder
	builder := NewBuilder(a).WithDatabase(a.database.GetPool())
	a.authService = builder.BuildAuthService()

	a.logger.Info("Services initialized")
	return nil
}

func (a *App) initTransports() error {
	// HTTP сервер
	a.httpServer = httpTransport.NewServer(
		a.config,
		a.authService,
		a.jwtService,
		a.validationService,
		a.logger,
		a.zapLogger,
	)

	// gRPC сервер
	a.grpcServer = grpcTransport.NewServer(
		a.config,
		a.authService,
		a.jwtService,
		a.validationService,
		a.logger,
	)

	a.logger.Info("Transport layers initialized")
	return nil
}

func (a *App) shutdown() error {
	a.logger.Info("Starting graceful shutdown")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	var shutdownErrors []error

	// Останавливаем HTTP сервер
	if a.httpServer != nil {
		if err := a.httpServer.Stop(shutdownCtx); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("HTTP server shutdown error: %w", err))
		}
	}

	// Останавливаем gRPC сервер
	if a.grpcServer != nil {
		if err := a.grpcServer.Stop(shutdownCtx); err != nil {
			shutdownErrors = append(shutdownErrors, fmt.Errorf("gRPC server shutdown error: %w", err))
		}
	}

	// Закрываем соединение с базой данных
	if a.database != nil {
		a.database.Close()
		a.logger.Info("Database connection closed")
	}

	if len(shutdownErrors) > 0 {
		for _, err := range shutdownErrors {
			a.logger.Error("Shutdown error", logger.Error(err))
		}
		return fmt.Errorf("shutdown completed with %d errors", len(shutdownErrors))
	}

	a.logger.Info("Graceful shutdown completed successfully")
	return nil
}

// Health проверяет состояние приложения
func (a *App) Health() map[string]interface{} {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"service":   a.config.Logger.ServiceName,
	}

	// Проверяем состояние компонентов
	components := make(map[string]string)

	// Проверка базы данных
	if a.database != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.database.HealthCheck(ctx); err != nil {
			components["database"] = "unhealthy"
		} else {
			components["database"] = "healthy"
		}
	} else {
		components["database"] = "not_initialized"
	}

	// Проверка серверов
	components["http_server"] = "healthy"
	components["grpc_server"] = "healthy"

	health["components"] = components

	// Определяем общий статус
	overallHealthy := true
	for _, status := range components {
		if status != "healthy" {
			overallHealthy = false
			break
		}
	}

	if !overallHealthy {
		health["status"] = "unhealthy"
	}

	return health
}

// GetConfig возвращает конфигурацию приложения
func (a *App) GetConfig() *config.Config {
	return a.config
}

// GetLogger возвращает логгер приложения
func (a *App) GetLogger() logger.Logger {
	return a.logger
}

// GetDB возвращает соединение с базой данных
func (a *App) GetDB() *pgxpool.Pool {
	if a.database == nil {
		return nil
	}
	return a.database.GetPool()
}

// GetDatabase возвращает объект базы данных
func (a *App) GetDatabase() *database.Database {
	return a.database
}
