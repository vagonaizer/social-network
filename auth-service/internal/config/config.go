package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Logger   LoggerConfig
}

type ServerConfig struct {
	HTTP HTTPConfig
	GRPC GRPCConfig
}

type HTTPConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type GRPCConfig struct {
	Port              string
	MaxReceiveSize    int
	MaxSendSize       int
	ConnectionTimeout time.Duration
	KeepaliveTime     time.Duration
	KeepaliveTimeout  time.Duration
}

type DatabaseConfig struct {
	Host              string
	Port              string
	User              string
	Password          string
	DBName            string
	SSLMode           string
	MaxConnections    int
	MinConnections    int
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout    time.Duration
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	Issuer        string
}

type LoggerConfig struct {
	Level       string
	ServiceName string
}

func Load() *Config {
	// 🟡 Попытка загрузить .env файл (мягко, не критично)
	if err := godotenv.Load(); err != nil {
		log.Println("[config] .env file not found or could not be loaded, continuing with system env")
	}

	return &Config{
		Server: ServerConfig{
			HTTP: HTTPConfig{
				Port:         getEnv("HTTP_PORT", "8080"),
				ReadTimeout:  getDurationEnv("HTTP_READ_TIMEOUT", 10*time.Second),
				WriteTimeout: getDurationEnv("HTTP_WRITE_TIMEOUT", 10*time.Second),
				IdleTimeout:  getDurationEnv("HTTP_IDLE_TIMEOUT", 60*time.Second),
			},
			GRPC: GRPCConfig{
				Port:              getEnv("GRPC_PORT", "9090"),
				MaxReceiveSize:    getIntEnv("GRPC_MAX_RECEIVE_SIZE", 4*1024*1024),
				MaxSendSize:       getIntEnv("GRPC_MAX_SEND_SIZE", 4*1024*1024),
				ConnectionTimeout: getDurationEnv("GRPC_CONNECTION_TIMEOUT", 5*time.Second),
				KeepaliveTime:     getDurationEnv("GRPC_KEEPALIVE_TIME", 30*time.Second),
				KeepaliveTimeout:  getDurationEnv("GRPC_KEEPALIVE_TIMEOUT", 5*time.Second),
			},
		},
		Database: DatabaseConfig{
			Host:              getEnv("DB_HOST", "localhost"),
			Port:              getEnv("DB_PORT", "5433"),
			User:              getEnv("DB_USER", "postgres"),
			Password:          getEnv("DB_PASSWORD", ""),
			DBName:            getEnv("DB_NAME", "auth_service"),
			SSLMode:           getEnv("DB_SSL_MODE", "disable"),
			MaxConnections:    getIntEnv("DB_MAX_CONNECTIONS", 30),
			MinConnections:    getIntEnv("DB_MIN_CONNECTIONS", 5),
			MaxConnLifetime:   getDurationEnv("DB_MAX_CONN_LIFETIME", time.Hour),
			MaxConnIdleTime:   getDurationEnv("DB_MAX_CONN_IDLE_TIME", 30*time.Minute),
			HealthCheckPeriod: getDurationEnv("DB_HEALTH_CHECK_PERIOD", time.Minute),
			ConnectTimeout:    getDurationEnv("DB_CONNECT_TIMEOUT", 10*time.Second),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "your-access-secret-key"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key"),
			Issuer:        getEnv("JWT_ISSUER", "auth-service"),
		},
		Logger: LoggerConfig{
			Level:       getEnv("LOG_LEVEL", "info"),
			ServiceName: getEnv("SERVICE_NAME", "auth-service"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
