package database

import (
	"context"
	"fmt"
	"social-network/auth-service/internal/config"
	"social-network/auth-service/pkg/logger"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool   *pgxpool.Pool
	config *config.DatabaseConfig
	logger logger.Logger
}

// NewDatabase создает новое подключение к базе данных используя конфигурацию
func NewDatabase(cfg *config.DatabaseConfig, logger logger.Logger) (*Database, error) {
	db := &Database{
		config: cfg,
		logger: logger,
	}

	if err := db.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}


func (d *Database) connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := d.buildDSN()

	d.logger.Info("Connecting to database",
		logger.String("host", d.config.Host),
		logger.String("port", d.config.Port),
		logger.String("database", d.config.DBName),
		logger.String("user", d.config.User),
		logger.String("ssl_mode", d.config.SSLMode),
	)

	// Парсим конфигурацию пула
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	// Настройки пула соединений из конфигурации
	poolConfig.MaxConns = int32(d.config.MaxConnections)
	poolConfig.MinConns = int32(d.config.MinConnections)
	poolConfig.MaxConnLifetime = d.config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = d.config.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = d.config.HealthCheckPeriod

	// Настройки таймаутов
	poolConfig.ConnConfig.ConnectTimeout = d.config.ConnectTimeout

	// Создаем пул соединений
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	d.pool = pool

	d.logger.Info("Database connection established successfully",
		logger.Int("max_connections", d.config.MaxConnections),
		logger.Int("min_connections", d.config.MinConnections),
		logger.String("max_conn_lifetime", d.config.MaxConnLifetime.String()),
	)

	return nil
}

// buildDSN строит строку подключения из конфигурации
func (d *Database) buildDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&pool_max_conns=%d&pool_min_conns=%d",
		d.config.User,
		d.config.Password,
		d.config.Host,
		d.config.Port,
		d.config.DBName,
		d.config.SSLMode,
		d.config.MaxConnections,
		d.config.MinConnections,
	)
}

// GetPool возвращает пул соединений
func (d *Database) GetPool() *pgxpool.Pool {
	return d.pool
}

// Close закрывает все соединения с базой данных
func (d *Database) Close() {
	if d.pool != nil {
		d.logger.Info("Closing database connections")
		d.pool.Close()
		d.logger.Info("Database connections closed")
	}
}

// Ping проверяет соединение с базой данных
func (d *Database) Ping(ctx context.Context) error {
	if d.pool == nil {
		return fmt.Errorf("database pool is not initialized")
	}

	return d.pool.Ping(ctx)
}

// Stats возвращает статистику пула соединений
func (d *Database) Stats() *pgxpool.Stat {
	if d.pool == nil {
		return nil
	}
	return d.pool.Stat()
}

// HealthCheck выполняет проверку здоровья базы данных
func (d *Database) HealthCheck(ctx context.Context) error {
	if d.pool == nil {
		return fmt.Errorf("database pool is not initialized")
	}

	// Создаем контекст с таймаутом для операций
	healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Проверяем ping
	if err := d.pool.Ping(healthCtx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Проверяем выполнение простого запроса
	var result int
	err := d.pool.QueryRow(healthCtx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database query test failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("database query returned unexpected result: %d", result)
	}

	return nil
}

// BeginTx начинает транзакцию
func (d *Database) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if d.pool == nil {
		return nil, fmt.Errorf("database pool is not initialized")
	}

	return d.pool.Begin(ctx)
}

// WithTx выполняет функцию в рамках транзакции
func (d *Database) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	if d.pool == nil {
		return fmt.Errorf("database pool is not initialized")
	}

	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			d.logger.Error("Failed to rollback transaction",
				logger.Error(rbErr),
				logger.Error(err),
			)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetConnectionInfo возвращает информацию о соединении
func (d *Database) GetConnectionInfo() map[string]interface{} {
	info := map[string]interface{}{
		"host":     d.config.Host,
		"port":     d.config.Port,
		"database": d.config.DBName,
		"user":     d.config.User,
		"ssl_mode": d.config.SSLMode,
	}

	if d.pool != nil {
		stats := d.pool.Stat()
		info["pool_stats"] = map[string]interface{}{
			"total_connections":        stats.TotalConns(),
			"idle_connections":         stats.IdleConns(),
			"acquired_connections":     stats.AcquiredConns(),
			"constructing_connections": stats.ConstructingConns(),
			"max_connections":          stats.MaxConns(),
		}
	}

	return info
}

func (d *Database) Stat() *pgxpool.Stat {
	if d.pool == nil {
		return &pgxpool.Stat{}
	}
	return d.pool.Stat()
}
