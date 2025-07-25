package handlers

import (
	"net/http"
	"social-network/auth-service/internal/domain"
	"social-network/auth-service/internal/service"
	"social-network/auth-service/internal/transport/http/dto"
	"social-network/auth-service/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService       *service.AuthService
	jwtService        *service.JWTService
	validationService *service.ValidationService
	logger            logger.Logger
}

func NewAuthHandler(
	authService *service.AuthService,
	jwtService *service.JWTService,
	validationService *service.ValidationService,
	logger logger.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService:       authService,
		jwtService:        jwtService,
		validationService: validationService,
		logger:            logger,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Registration data"
// @Success 201 {object} dto.RegisterResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Валидация данных
	if err := h.validationService.ValidateRegistrationData(
		req.Email, req.Username, req.DisplayName, req.Password,
	); err != nil {
		h.respondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Регистрация пользователя
	user, err := h.authService.RegisterUser(req.Email, req.Username, req.DisplayName, req.Password)
	if err != nil {
		h.logger.Error("Registration failed",
			logger.String("email", req.Email),
			logger.String("username", req.Username),
			logger.Error(err),
		)
		h.handleServiceError(c, err)
		return
	}

	response := dto.RegisterResponse{
		User:    h.mapUserToDTO(user),
		Message: "User registered successfully. Please check your email for verification.",
	}

	h.logger.Info("User registered successfully",
		logger.String("user_id", user.ID().String()),
		logger.String("username", user.Username()),
	)

	c.JSON(http.StatusCreated, response)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Аутентификация
	user, err := h.authService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		h.logger.Warn("Login attempt failed",
			logger.String("email", req.Email),
			logger.String("client_ip", c.ClientIP()),
			logger.Error(err),
		)
		h.handleServiceError(c, err)
		return
	}

	// Получение ролей
	roles, err := h.authService.GetUserRoles(user.ID())
	if err != nil {
		h.logger.Error("Failed to get user roles",
			logger.String("user_id", user.ID().String()),
			logger.Error(err),
		)
		roles = []*domain.UserRole{} // Пустой массив ролей
	}

	// Генерация токенов
	roleStrings := make([]domain.UserRoleType, len(roles))
	for i, role := range roles {
		roleStrings[i] = role.Role()
	}

	accessToken, err := h.jwtService.GenerateAccessToken(user, roleStrings)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, "token_generation_error", "Failed to generate access token")
		return
	}

	refreshTokenEntity, err := h.authService.CreateRefreshToken(user.ID())
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, "token_generation_error", "Failed to generate refresh token")
		return
	}

	response := dto.LoginResponse{
		Tokens: dto.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshTokenEntity.Token(),
			TokenType:    "Bearer",
			ExpiresIn:    900, // 15 minutes
		},
		User: h.mapUserToDTO(user),
	}

	h.logger.Info("User logged in successfully",
		logger.String("user_id", user.ID().String()),
		logger.String("username", user.Username()),
		logger.String("client_ip", c.ClientIP()),
	)

	c.JSON(http.StatusOK, response)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Generate new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} dto.TokenResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	// Валидация refresh token
	refreshToken, err := h.authService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "invalid_token", "Invalid or expired refresh token")
		return
	}

	// Получение пользователя
	user, err := h.authService.GetUserByID(refreshToken.UserID())
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "user_not_found", "User not found")
		return
	}

	// Получение ролей
	roles, err := h.authService.GetUserRoles(user.ID())
	if err != nil {
		roles = []*domain.UserRole{}
	}

	roleStrings := make([]domain.UserRoleType, len(roles))
	for i, role := range roles {
		roleStrings[i] = role.Role()
	}

	// Генерация нового access token
	accessToken, err := h.jwtService.GenerateAccessToken(user, roleStrings)
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, "token_generation_error", "Failed to generate access token")
		return
	}

	// Создание нового refresh token
	newRefreshToken, err := h.authService.CreateRefreshToken(user.ID())
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, "token_generation_error", "Failed to generate refresh token")
		return
	}

	// Отзыв старого refresh token
	if err := h.authService.RevokeRefreshToken(req.RefreshToken); err != nil {
		h.logger.Error("Failed to revoke old refresh token",
			logger.String("user_id", user.ID().String()),
			logger.Error(err),
		)
	}

	response := dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken.Token(),
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 minutes
	}

	c.JSON(http.StatusOK, response)
}

// VerifyEmail godoc
// @Summary Verify email address
// @Description Verify user's email address using verification token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.VerifyEmailRequest true "Verification token"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req dto.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	if err := h.authService.VerifyEmail(req.Token); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Email verified successfully",
	})
}

// InitiatePasswordReset godoc
// @Summary Initiate password reset
// @Description Send password reset email to user
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.InitiatePasswordResetRequest true "Email address"
// @Success 200 {object} dto.MessageResponse
// @Router /auth/reset-password [post]
func (h *AuthHandler) InitiatePasswordReset(c *gin.Context) {
	var req dto.InitiatePasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	if err := h.authService.InitiatePasswordReset(req.Email); err != nil {
		h.logger.Error("Password reset initiation failed",
			logger.String("email", req.Email),
			logger.Error(err),
		)
	}

	// Всегда возвращаем успешный ответ для безопасности
	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset user password using reset token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /auth/reset-password/confirm [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req dto.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	if err := h.validationService.ValidatePassword(req.NewPassword); err != nil {
		h.respondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Password reset successfully",
	})
}

// GetCurrentUser godoc
// @Summary Get current user
// @Description Get current authenticated user information
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondError(c, http.StatusUnauthorized, "unauthorized", "User not authenticated")
		return
	}

	user, err := h.authService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, h.mapUserToDTO(user))
}

// ChangePassword godoc
// @Summary Change password
// @Description Change user password
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ChangePasswordRequest true "Current and new password"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/change-password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondError(c, http.StatusUnauthorized, "unauthorized", "User not authenticated")
		return
	}

	var req dto.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	if err := h.validationService.ValidatePassword(req.NewPassword); err != nil {
		h.respondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	if err := h.authService.ChangePassword(userID.(uuid.UUID), req.CurrentPassword, req.NewPassword); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Password changed successfully",
	})
}

// Logout godoc
// @Summary Logout user
// @Description Logout user and revoke refresh token
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.LogoutRequest true "Refresh token"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	if err := h.authService.RevokeRefreshToken(req.RefreshToken); err != nil {
		h.logger.Error("Failed to revoke refresh token during logout",
			logger.String("token", req.RefreshToken),
			logger.Error(err),
		)
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Logged out successfully",
	})
}

// ValidateToken godoc
// @Summary Validate access token
// @Description Validate access token and return user info (internal use)
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.ValidateTokenResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /auth/validate [get]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondError(c, http.StatusUnauthorized, "unauthorized", "Invalid token")
		return
	}

	user, err := h.authService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		h.respondError(c, http.StatusUnauthorized, "unauthorized", "User not found")
		return
	}

	roles, exists := c.Get("user_roles")
	roleStrings := []string{}
	if exists {
		if userRoles, ok := roles.([]domain.UserRoleType); ok {
			for _, role := range userRoles {
				roleStrings = append(roleStrings, string(role))
			}
		}
	}

	response := dto.ValidateTokenResponse{
		Valid: true,
		User:  h.mapUserToDTO(user),
		Roles: roleStrings,
	}

	c.JSON(http.StatusOK, response)
}

// AssignRole godoc
// @Summary Assign role to user
// @Description Assign a role to a user (admin only)
// @Tags admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param request body dto.AssignRoleRequest true "Role to assign"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /auth/users/{user_id}/roles [post]
func (h *AuthHandler) AssignRole(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "validation_error", "Invalid user ID")
		return
	}

	var req dto.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	roleType := domain.UserRoleType(req.Role)
	if err := h.authService.AssignRole(userID, roleType); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Role assigned successfully",
	})
}

// RevokeRole godoc
// @Summary Revoke role from user
// @Description Revoke a role from a user (admin only)
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param user_id path string true "User ID"
// @Param role path string true "Role to revoke"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /auth/users/{user_id}/roles/{role} [delete]
func (h *AuthHandler) RevokeRole(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "validation_error", "Invalid user ID")
		return
	}

	role := c.Param("role")
	roleType := domain.UserRoleType(role)

	if err := h.authService.RevokeRole(userID, roleType); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.MessageResponse{
		Message: "Role revoked successfully",
	})
}

// GetUserRoles godoc
// @Summary Get user roles
// @Description Get all roles assigned to a user (admin only)
// @Tags admin
// @Security BearerAuth
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} dto.GetUserRolesResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Router /auth/users/{user_id}/roles [get]
func (h *AuthHandler) GetUserRoles(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "validation_error", "Invalid user ID")
		return
	}

	roles, err := h.authService.GetUserRoles(userID)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	roleResponses := make([]dto.UserRoleResponse, len(roles))
	for i, role := range roles {
		roleResponses[i] = dto.UserRoleResponse{
			ID:        role.ID(),
			UserID:    role.UserID(),
			Role:      string(role.Role()),
			GrantedAt: role.GrantedAt(),
			IsActive:  role.IsActive(),
		}
	}

	response := dto.GetUserRolesResponse{
		Roles: roleResponses,
	}

	c.JSON(http.StatusOK, response)
}

// Helper methods
func (h *AuthHandler) mapUserToDTO(user *domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:          user.ID(),
		Email:       user.Email(),
		Username:    user.Username(),
		DisplayName: user.DisplayName(),
		IsVerified:  user.IsVerified(),
		IsActive:    user.IsActive(),
		CreatedAt:   user.CreatedAt(),
		UpdatedAt:   user.UpdatedAt(),
	}
}

func (h *AuthHandler) respondError(c *gin.Context, statusCode int, errorType, message string) {
	c.JSON(statusCode, dto.ErrorResponse{
		Error:     errorType,
		Message:   message,
		Timestamp: time.Now(),
		Path:      c.Request.URL.Path,
	})
}

func (h *AuthHandler) handleServiceError(c *gin.Context, err error) {
	// Здесь можно добавить более детальную обработку различных типов ошибок
	switch err.Error() {
	case "user not found":
		h.respondError(c, http.StatusNotFound, "user_not_found", "User not found")
	case "invalid email or password":
		h.respondError(c, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password")
	case "user account is inactive":
		h.respondError(c, http.StatusForbidden, "account_inactive", "User account is inactive")
	case "email is already verified":
		h.respondError(c, http.StatusBadRequest, "already_verified", "Email is already verified")
	case "user with this email already exists":
		h.respondError(c, http.StatusConflict, "email_exists", "User with this email already exists")
	case "user with this username already exists":
		h.respondError(c, http.StatusConflict, "username_exists", "User with this username already exists")
	case "email verification not found":
		h.respondError(c, http.StatusNotFound, "verification_not_found", "Email verification token not found")
	case "email verification token has expired":
		h.respondError(c, http.StatusBadRequest, "verification_expired", "Email verification token has expired")
	case "email verification token has already been used":
		h.respondError(c, http.StatusBadRequest, "verification_used", "Email verification token has already been used")
	case "email verification token is invalid":
		h.respondError(c, http.StatusBadRequest, "verification_invalid", "Email verification token is invalid")
	default:
		h.logger.Error("Unhandled service error", logger.Error(err))
		h.respondError(c, http.StatusInternalServerError, "internal_error", "Internal server error")
	}
}
