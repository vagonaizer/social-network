package service

import (
	"social-network/auth-service/internal/domain"
	"social-network/auth-service/internal/repository"
	"social-network/auth-service/pkg/helpers"
	"social-network/auth-service/pkg/logger"
	"time"

	"github.com/google/uuid"
)

type AuthService struct {
	userRepo              repository.UserRepository
	userAuthRepo          repository.UserAuthRepository
	userRoleRepo          repository.UserRoleRepository
	refreshTokenRepo      repository.RefreshTokenRepository
	emailVerificationRepo repository.EmailVerificationRepository
	passwordResetRepo     repository.PasswordResetRepository
	logger                logger.Logger
}

func NewAuthService(
	userRepo repository.UserRepository,
	userAuthRepo repository.UserAuthRepository,
	userRoleRepo repository.UserRoleRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	emailVerificationRepo repository.EmailVerificationRepository,
	passwordResetRepo repository.PasswordResetRepository,
	logger logger.Logger,
) *AuthService {
	return &AuthService{
		userRepo:              userRepo,
		userAuthRepo:          userAuthRepo,
		userRoleRepo:          userRoleRepo,
		refreshTokenRepo:      refreshTokenRepo,
		emailVerificationRepo: emailVerificationRepo,
		passwordResetRepo:     passwordResetRepo,
		logger:                logger,
	}
}

// GetUserByID получает пользователя по ID
func (s *AuthService) GetUserByID(userID uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(userID)
}

// RegisterUser создает нового пользователя
func (s *AuthService) RegisterUser(email, username, displayName, password string) (*domain.User, error) {
	s.logger.Info("Starting user registration",
		logger.String("email", email),
		logger.String("username", username),
	)

	// Проверяем существование пользователя
	if exists, err := s.userRepo.ExistsByEmail(email); err != nil {
		return nil, err
	} else if exists {
		return nil, repository.ErrUserEmailExists
	}

	if exists, err := s.userRepo.ExistsByUsername(username); err != nil {
		return nil, err
	} else if exists {
		return nil, repository.ErrUserUsernameExists
	}

	// Создаем пользователя
	user := domain.NewUser(email, username, displayName)
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Хешируем пароль и создаем auth запись
	hashedPassword, err := helpers.HashPassword(password)
	if err != nil {
		return nil, err
	}

	userAuth := domain.NewUserAuth(user.ID(), hashedPassword)
	if err := s.userAuthRepo.Create(userAuth); err != nil {
		return nil, err
	}

	// Назначаем базовую роль пользователя
	userRole := domain.NewUserRole(user.ID(), domain.RoleUser)
	if err := s.userRoleRepo.Create(userRole); err != nil {
		return nil, err
	}

	// Создаем токен для верификации email
	if err := s.createEmailVerification(user.ID()); err != nil {
		s.logger.Error("Failed to create email verification",
			logger.String("user_id", user.ID().String()),
			logger.Error(err),
		)
	}

	s.logger.Info("User registered successfully",
		logger.String("user_id", user.ID().String()),
		logger.String("username", username),
	)

	return user, nil
}

// AuthenticateUser проверяет учетные данные пользователя
func (s *AuthService) AuthenticateUser(email, password string) (*domain.User, error) {
	// Получаем пользователя
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, repository.ErrUserNotFound
	}

	// Проверяем активность аккаунта
	if !user.IsActive() {
		return nil, ErrUserInactive
	}

	// Получаем данные аутентификации
	userAuth, err := s.userAuthRepo.GetByUserID(user.ID())
	if err != nil {
		return nil, repository.ErrUserAuthNotFound
	}

	// Проверяем пароль
	if err := helpers.ComparePassword(userAuth.PasswordHash(), password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Обновляем время последнего входа
	now := time.Now()
	userAuth.SetLastLoginAt(&now)
	if err := s.userAuthRepo.Update(userAuth); err != nil {
		s.logger.Error("Failed to update last login time",
			logger.String("user_id", user.ID().String()),
			logger.Error(err),
		)
	}

	return user, nil
}

// CreateRefreshToken создает refresh token для пользователя
func (s *AuthService) CreateRefreshToken(userID uuid.UUID) (*domain.RefreshToken, error) {
	token := helpers.GenerateSecureToken()
	expiresAt := helpers.GetExpirationTime("refresh")

	refreshToken := domain.NewRefreshToken(userID, token, expiresAt)
	if err := s.refreshTokenRepo.Create(refreshToken); err != nil {
		return nil, err
	}

	return refreshToken, nil
}

// ValidateRefreshToken проверяет refresh token
func (s *AuthService) ValidateRefreshToken(token string) (*domain.RefreshToken, error) {
	refreshToken, err := s.refreshTokenRepo.GetByToken(token)
	if err != nil {
		return nil, err
	}

	if !refreshToken.IsValid() {
		return nil, repository.ErrRefreshTokenInvalid
	}

	return refreshToken, nil
}

// RevokeRefreshToken отзывает refresh token
func (s *AuthService) RevokeRefreshToken(token string) error {
	refreshToken, err := s.refreshTokenRepo.GetByToken(token)
	if err != nil {
		return err
	}

	refreshToken.SetRevoked(true)
	return s.refreshTokenRepo.Update(refreshToken)
}

// RevokeAllUserTokens отзывает все refresh токены пользователя
func (s *AuthService) RevokeAllUserTokens(userID uuid.UUID) error {
	tokens, err := s.refreshTokenRepo.GetByUserID(userID)
	if err != nil {
		return err
	}

	for _, token := range tokens {
		if !token.IsRevoked() {
			token.SetRevoked(true)
			if err := s.refreshTokenRepo.Update(token); err != nil {
				s.logger.Error("Failed to revoke refresh token",
					logger.String("user_id", userID.String()),
					logger.String("token", token.Token()),
					logger.Error(err),
				)
			}
		}
	}

	return nil
}

// VerifyEmail подтверждает email пользователя
func (s *AuthService) VerifyEmail(token string) error {
	verification, err := s.emailVerificationRepo.GetByToken(token)
	if err != nil {
		return err
	}

	if !verification.IsValid() {
		return repository.ErrEmailVerificationInvalid
	}

	// Получаем пользователя
	user, err := s.userRepo.GetByID(verification.UserID())
	if err != nil {
		return err
	}

	// Проверяем, не верифицирован ли уже
	if user.IsVerified() {
		return ErrEmailAlreadyVerified
	}

	// Верифицируем пользователя
	user.SetVerified(true)
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	// Помечаем токен как использованный
	verification.SetUsed(true)
	return s.emailVerificationRepo.Update(verification)
}

// InitiatePasswordReset создает токен для сброса пароля
func (s *AuthService) InitiatePasswordReset(email string) error {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		// Не раскрываем информа��ию о существовании email
		return nil
	}

	if !user.IsActive() {
		return nil
	}

	return s.createPasswordReset(user.ID())
}

// ResetPassword сбрасывает пароль пользователя
func (s *AuthService) ResetPassword(token, newPassword string) error {
	reset, err := s.passwordResetRepo.GetByToken(token)
	if err != nil {
		return err
	}

	if !reset.IsValid() {
		return repository.ErrPasswordResetInvalid
	}

	// Хешируем новый пароль
	hashedPassword, err := helpers.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Обновляем пароль
	userAuth, err := s.userAuthRepo.GetByUserID(reset.UserID())
	if err != nil {
		return err
	}

	userAuth.SetPasswordHash(hashedPassword)
	if err := s.userAuthRepo.Update(userAuth); err != nil {
		return err
	}

	// Помечаем токен как использованный
	reset.SetUsed(true)
	if err := s.passwordResetRepo.Update(reset); err != nil {
		return err
	}

	// Отзываем все refresh токены пользователя
	return s.RevokeAllUserTokens(reset.UserID())
}

// ChangePassword изменяет пароль пользователя
func (s *AuthService) ChangePassword(userID uuid.UUID, currentPassword, newPassword string) error {
	userAuth, err := s.userAuthRepo.GetByUserID(userID)
	if err != nil {
		return err
	}

	// Проверяем текущий пароль
	if err := helpers.ComparePassword(userAuth.PasswordHash(), currentPassword); err != nil {
		return ErrInvalidCurrentPassword
	}

	// Хешируем новый пароль
	hashedPassword, err := helpers.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Обновляем пароль
	userAuth.SetPasswordHash(hashedPassword)
	if err := s.userAuthRepo.Update(userAuth); err != nil {
		return err
	}

	// Отзываем все refresh токены кроме текущего
	return s.RevokeAllUserTokens(userID)
}

// GetUserRoles возвращает роли пользователя
func (s *AuthService) GetUserRoles(userID uuid.UUID) ([]*domain.UserRole, error) {
	return s.userRoleRepo.GetByUserID(userID)
}

// HasRole проверяет наличие роли у пользователя
func (s *AuthService) HasRole(userID uuid.UUID, role domain.UserRoleType) (bool, error) {
	roles, err := s.userRoleRepo.GetByUserID(userID)
	if err != nil {
		return false, err
	}

	for _, userRole := range roles {
		if userRole.Role() == role && userRole.IsActive() {
			return true, nil
		}
	}

	return false, nil
}

// HasPermission проверяет, имеет ли пользователь необходимые права
func (s *AuthService) HasPermission(userID uuid.UUID, requiredRole domain.UserRoleType) (bool, error) {
	roles, err := s.userRoleRepo.GetByUserID(userID)
	if err != nil {
		return false, err
	}

	for _, userRole := range roles {
		if userRole.IsActive() && helpers.HasHigherRole(userRole.Role(), requiredRole) {
			return true, nil
		}
	}

	return false, nil
}

// AssignRole назначает роль пользователю
func (s *AuthService) AssignRole(userID uuid.UUID, role domain.UserRoleType) error {
	// Проверяем валидность роли
	if !helpers.IsValidRole(role) {
		return repository.ErrInvalidRole
	}

	// Проверяем, нет ли уже такой роли
	hasRole, err := s.HasRole(userID, role)
	if err != nil {
		return err
	}

	if hasRole {
		return repository.ErrUserRoleAlreadyExists
	}

	userRole := domain.NewUserRole(userID, role)
	return s.userRoleRepo.Create(userRole)
}

// RevokeRole отзывает роль у пользователя
func (s *AuthService) RevokeRole(userID uuid.UUID, role domain.UserRoleType) error {
	roles, err := s.userRoleRepo.GetByUserID(userID)
	if err != nil {
		return err
	}

	for _, userRole := range roles {
		if userRole.Role() == role && userRole.IsActive() {
			userRole.SetActive(false)
			return s.userRoleRepo.Update(userRole)
		}
	}

	return repository.ErrUserRoleNotFound
}

// Приватные методы

func (s *AuthService) createEmailVerification(userID uuid.UUID) error {
	token := helpers.GenerateSecureToken()
	expiresAt := helpers.GetExpirationTime("email_verification")

	verification := domain.NewEmailVerification(userID, token, expiresAt)
	return s.emailVerificationRepo.Create(verification)
}

func (s *AuthService) createPasswordReset(userID uuid.UUID) error {
	token := helpers.GenerateSecureToken()
	expiresAt := helpers.GetExpirationTime("password_reset")

	reset := domain.NewPasswordReset(userID, token, expiresAt)
	return s.passwordResetRepo.Create(reset)
}
