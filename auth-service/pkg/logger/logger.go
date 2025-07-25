package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger интерфейс для логирования
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

// Field представляет поле для логирования
type Field struct {
	Key   string
	Value interface{}
}

// Функции для создания полей
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// CustomLogger наш кастомный логгер для обычных операций
type CustomLogger struct {
	logger *slog.Logger
	fields []Field
}

// ZapLogger обертка для zap (для HTTP запросов)
type ZapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// NewCustomLogger создает новый кастомный логгер
func NewCustomLogger(serviceName string, level string, output io.Writer) *CustomLogger {
	if output == nil {
		output = os.Stdout
	}

	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Кастомизируем формат времени
			if a.Key == slog.TimeKey {
				return slog.String("timestamp", a.Value.Time().Format(time.RFC3339))
			}
			// Упрощаем путь к файлу
			if a.Key == slog.SourceKey {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					source.File = getShortFileName(source.File)
				}
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(output, opts)
	logger := slog.New(handler).With(
		slog.String("service", serviceName),
		slog.String("version", "1.0.0"),
	)

	return &CustomLogger{
		logger: logger,
		fields: make([]Field, 0),
	}
}

// NewZapLogger создает новый zap логгер для HTTP запросов
func NewZapLogger(serviceName string, level string) *ZapLogger {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields: map[string]interface{}{
			"service": serviceName,
			"version": "1.0.0",
		},
	}

	logger, _ := config.Build()
	return &ZapLogger{
		logger: logger,
		sugar:  logger.Sugar(),
	}
}

// Методы CustomLogger
func (l *CustomLogger) Debug(msg string, fields ...Field) {
	l.log(slog.LevelDebug, msg, fields...)
}

func (l *CustomLogger) Info(msg string, fields ...Field) {
	l.log(slog.LevelInfo, msg, fields...)
}

func (l *CustomLogger) Warn(msg string, fields ...Field) {
	l.log(slog.LevelWarn, msg, fields...)
}

func (l *CustomLogger) Error(msg string, fields ...Field) {
	l.log(slog.LevelError, msg, fields...)
}

func (l *CustomLogger) Fatal(msg string, fields ...Field) {
	l.log(slog.LevelError, msg, fields...)
	os.Exit(1)
}

func (l *CustomLogger) With(fields ...Field) Logger {
	return &CustomLogger{
		logger: l.logger,
		fields: append(l.fields, fields...),
	}
}

func (l *CustomLogger) WithContext(ctx context.Context) Logger {
	// Можно добавить извлечение данных из контекста (trace_id, user_id и т.д.)
	return l
}

func (l *CustomLogger) log(level slog.Level, msg string, fields ...Field) {
	allFields := append(l.fields, fields...)
	attrs := make([]slog.Attr, 0, len(allFields))

	for _, field := range allFields {
		attrs = append(attrs, slog.Any(field.Key, field.Value))
	}

	l.logger.LogAttrs(context.Background(), level, msg, attrs...)
}

// Методы ZapLogger
func (z *ZapLogger) Debug(msg string, fields ...Field) {
	z.logger.Debug(msg, z.convertFields(fields...)...)
}

func (z *ZapLogger) Info(msg string, fields ...Field) {
	z.logger.Info(msg, z.convertFields(fields...)...)
}

func (z *ZapLogger) Warn(msg string, fields ...Field) {
	z.logger.Warn(msg, z.convertFields(fields...)...)
}

func (z *ZapLogger) Error(msg string, fields ...Field) {
	z.logger.Error(msg, z.convertFields(fields...)...)
}

func (z *ZapLogger) Fatal(msg string, fields ...Field) {
	z.logger.Fatal(msg, z.convertFields(fields...)...)
}

func (z *ZapLogger) With(fields ...Field) Logger {
	return &ZapLogger{
		logger: z.logger.With(z.convertFields(fields...)...),
		sugar:  z.sugar,
	}
}

func (z *ZapLogger) WithContext(ctx context.Context) Logger {
	return z
}

func (z *ZapLogger) convertFields(fields ...Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		switch v := field.Value.(type) {
		case string:
			zapFields = append(zapFields, zap.String(field.Key, v))
		case int:
			zapFields = append(zapFields, zap.Int(field.Key, v))
		case int64:
			zapFields = append(zapFields, zap.Int64(field.Key, v))
		case float64:
			zapFields = append(zapFields, zap.Float64(field.Key, v))
		case bool:
			zapFields = append(zapFields, zap.Bool(field.Key, v))
		case error:
			zapFields = append(zapFields, zap.Error(v))
		case time.Duration:
			zapFields = append(zapFields, zap.Duration(field.Key, v))
		default:
			zapFields = append(zapFields, zap.Any(field.Key, v))
		}
	}
	return zapFields
}

// Sugar возвращает sugared logger для более простого использования
func (z *ZapLogger) Sugar() *zap.SugaredLogger {
	return z.sugar
}

// Вспомогательные функции
func getShortFileName(fullPath string) string {
	for i := len(fullPath) - 1; i > 0; i-- {
		if fullPath[i] == '/' {
			return fullPath[i+1:]
		}
	}
	return fullPath
}

// GetCaller возвращает информацию о вызывающем коде
func GetCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "unknown"
	}
	return fmt.Sprintf("%s:%d", getShortFileName(file), line)
}
