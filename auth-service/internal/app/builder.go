package app

import (
	"social-network/auth-service/internal/infrastructure/postgres"
	"social-network/auth-service/internal/repository"
	"social-network/auth-service/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Builder помогает собирать зависимости приложения
type Builder struct {
	app *App
	db  *pgxpool.Pool
}

// NewBuilder создает новый builder
func NewBuilder(app *App) *Builder {
	return &Builder{
		app: app,
	}
}

// WithDatabase устанавливает соединение с базой данных
func (b *Builder) WithDatabase(db *pgxpool.Pool) *Builder {
	b.db = db
	return b
}

// BuildRepositories создает все репозитории
func (b *Builder) BuildRepositories() (
	repository.UserRepository,
	repository.UserAuthRepository,
	repository.UserRoleRepository,
	repository.RefreshTokenRepository,
	repository.EmailVerificationRepository,
	repository.PasswordResetRepository,
) {
	userRepo := postgres.NewUserRepository(b.db)
	userAuthRepo := postgres.NewUserAuthRepository(b.db)
	userRoleRepo := postgres.NewUserRoleRepository(b.db)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(b.db)
	emailVerificationRepo := postgres.NewEmailVerificationRepository(b.db)
	passwordResetRepo := postgres.NewPasswordResetRepository(b.db)

	return userRepo, userAuthRepo, userRoleRepo, refreshTokenRepo, emailVerificationRepo, passwordResetRepo
}

// BuildAuthService создает сервис аутентификации
func (b *Builder) BuildAuthService() *service.AuthService {
	userRepo, userAuthRepo, userRoleRepo, refreshTokenRepo, emailVerificationRepo, passwordResetRepo := b.BuildRepositories()

	return service.NewAuthService(
		userRepo,
		userAuthRepo,
		userRoleRepo,
		refreshTokenRepo,
		emailVerificationRepo,
		passwordResetRepo,
		b.app.logger,
	)
}
