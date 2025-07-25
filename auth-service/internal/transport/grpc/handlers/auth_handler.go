package handlers

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"social-network/auth-service/internal/domain"
	"social-network/auth-service/internal/service"
	pb "social-network/auth-service/pkg/api/proto/auth/v1"
	"social-network/auth-service/pkg/logger"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
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

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Валидация данных
	if err := h.validationService.ValidateRegistrationData(
		req.Email, req.Username, req.DisplayName, req.Password,
	); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
	}

	// Регистрация пользователя
	user, err := h.authService.RegisterUser(req.Email, req.Username, req.DisplayName, req.Password)
	if err != nil {
		h.logger.Error("Registration failed",
			logger.String("email", req.Email),
			logger.String("username", req.Username),
			logger.Error(err),
		)
		return nil, h.handleServiceError(err)
	}

	h.logger.Info("User registered successfully",
		logger.String("user_id", user.ID().String()),
		logger.String("username", user.Username()),
	)

	return &pb.RegisterResponse{
		User:    h.mapUserToPB(user),
		Message: "User registered successfully. Please check your email for verification.",
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Аутентификация
	user, err := h.authService.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		h.logger.Warn("Login attempt failed",
			logger.String("email", req.Email),
			logger.Error(err),
		)
		return nil, h.handleServiceError(err)
	}

	// Получение ролей
	roles, err := h.authService.GetUserRoles(user.ID())
	if err != nil {
		h.logger.Error("Failed to get user roles",
			logger.String("user_id", user.ID().String()),
			logger.Error(err),
		)
		roles = []*domain.UserRole{}
	}

	// Генерация токенов
	roleStrings := make([]domain.UserRoleType, len(roles))
	for i, role := range roles {
		roleStrings[i] = role.Role()
	}

	accessToken, err := h.jwtService.GenerateAccessToken(user, roleStrings)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate access token")
	}

	refreshTokenEntity, err := h.authService.CreateRefreshToken(user.ID())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token")
	}

	h.logger.Info("User logged in successfully",
		logger.String("user_id", user.ID().String()),
		logger.String("username", user.Username()),
	)

	return &pb.LoginResponse{
		Tokens: &pb.TokenPair{
			AccessToken:  accessToken,
			RefreshToken: refreshTokenEntity.Token(),
			TokenType:    "Bearer",
			ExpiresIn:    900, // 15 minutes
		},
		User: h.mapUserToPB(user),
	}, nil
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	// Валидация refresh token
	refreshToken, err := h.authService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid or expired refresh token")
	}

	// Получение пользователя
	user, err := h.authService.GetUserByID(refreshToken.UserID())
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "user not found")
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
		return nil, status.Errorf(codes.Internal, "failed to generate access token")
	}

	// Создание нового refresh token
	newRefreshToken, err := h.authService.CreateRefreshToken(user.ID())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate refresh token")
	}

	// Отзыв старого refresh token
	if err := h.authService.RevokeRefreshToken(req.RefreshToken); err != nil {
		h.logger.Error("Failed to revoke old refresh token",
			logger.String("user_id", user.ID().String()),
			logger.Error(err),
		)
	}

	return &pb.RefreshTokenResponse{
		Tokens: &pb.TokenPair{
			AccessToken:  accessToken,
			RefreshToken: newRefreshToken.Token(),
			TokenType:    "Bearer",
			ExpiresIn:    900, // 15 minutes
		},
	}, nil
}

func (h *AuthHandler) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	if err := h.authService.VerifyEmail(req.Token); err != nil {
		return nil, h.handleServiceError(err)
	}

	return &pb.VerifyEmailResponse{
		Message: "Email verified successfully",
	}, nil
}

func (h *AuthHandler) InitiatePasswordReset(ctx context.Context, req *pb.InitiatePasswordResetRequest) (*pb.InitiatePasswordResetResponse, error) {
	if err := h.authService.InitiatePasswordReset(req.Email); err != nil {
		h.logger.Error("Password reset initiation failed",
			logger.String("email", req.Email),
			logger.Error(err),
		)
	}

	// Всегда возвращаем успешный ответ для безопасности
	return &pb.InitiatePasswordResetResponse{
		Message: "If the email exists, a password reset link has been sent",
	}, nil
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	if err := h.validationService.ValidatePassword(req.NewPassword); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
	}

	if err := h.authService.ResetPassword(req.Token, req.NewPassword); err != nil {
		return nil, h.handleServiceError(err)
	}

	return &pb.ResetPasswordResponse{
		Message: "Password reset successfully",
	}, nil
}

func (h *AuthHandler) GetCurrentUser(ctx context.Context, req *pb.GetCurrentUserRequest) (*pb.GetCurrentUserResponse, error) {
	// Валидация токена
	claims, err := h.jwtService.ValidateAccessToken(req.AccessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid or expired token")
	}

	user, err := h.authService.GetUserByID(claims.UserID)
	if err != nil {
		return nil, h.handleServiceError(err)
	}

	return &pb.GetCurrentUserResponse{
		User: h.mapUserToPB(user),
	}, nil
}

func (h *AuthHandler) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	// Валидация токена
	claims, err := h.jwtService.ValidateAccessToken(req.AccessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid or expired token")
	}

	if err := h.validationService.ValidatePassword(req.NewPassword); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "validation error: %v", err)
	}

	if err := h.authService.ChangePassword(claims.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		return nil, h.handleServiceError(err)
	}

	return &pb.ChangePasswordResponse{
		Message: "Password changed successfully",
	}, nil
}

func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	// Валидация токена
	_, err := h.jwtService.ValidateAccessToken(req.AccessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid or expired token")
	}

	if err := h.authService.RevokeRefreshToken(req.RefreshToken); err != nil {
		h.logger.Error("Failed to revoke refresh token during logout",
			logger.String("token", req.RefreshToken),
			logger.Error(err),
		)
	}

	return &pb.LogoutResponse{
		Message: "Logged out successfully",
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	claims, err := h.jwtService.ValidateAccessToken(req.AccessToken)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	user, err := h.authService.GetUserByID(claims.UserID)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid: false,
		}, nil
	}

	roleStrings := make([]string, len(claims.Roles))
	for i, role := range claims.Roles {
		roleStrings[i] = string(role)
	}

	return &pb.ValidateTokenResponse{
		Valid: true,
		User:  h.mapUserToPB(user),
		Roles: roleStrings,
	}, nil
}

func (h *AuthHandler) AssignRole(ctx context.Context, req *pb.AssignRoleRequest) (*pb.AssignRoleResponse, error) {
	// Валидация токена и проверка прав администратора
	claims, err := h.jwtService.ValidateAccessToken(req.AccessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid or expired token")
	}

	hasPermission := false
	for _, role := range claims.Roles {
		if role == domain.RoleAdmin {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	roleType := domain.UserRoleType(req.Role)
	if err := h.authService.AssignRole(userID, roleType); err != nil {
		return nil, h.handleServiceError(err)
	}

	return &pb.AssignRoleResponse{
		Message: "Role assigned successfully",
	}, nil
}

func (h *AuthHandler) RevokeRole(ctx context.Context, req *pb.RevokeRoleRequest) (*pb.RevokeRoleResponse, error) {
	// Валидация токена и проверка прав администратора
	claims, err := h.jwtService.ValidateAccessToken(req.AccessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid or expired token")
	}

	hasPermission := false
	for _, role := range claims.Roles {
		if role == domain.RoleAdmin {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	roleType := domain.UserRoleType(req.Role)
	if err := h.authService.RevokeRole(userID, roleType); err != nil {
		return nil, h.handleServiceError(err)
	}

	return &pb.RevokeRoleResponse{
		Message: "Role revoked successfully",
	}, nil
}

func (h *AuthHandler) GetUserRoles(ctx context.Context, req *pb.GetUserRolesRequest) (*pb.GetUserRolesResponse, error) {
	// Валидация токена и проверка прав администратора
	claims, err := h.jwtService.ValidateAccessToken(req.AccessToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid or expired token")
	}

	hasPermission := false
	for _, role := range claims.Roles {
		if role == domain.RoleAdmin {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return nil, status.Errorf(codes.PermissionDenied, "insufficient permissions")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user ID")
	}

	roles, err := h.authService.GetUserRoles(userID)
	if err != nil {
		return nil, h.handleServiceError(err)
	}

	pbRoles := make([]*pb.UserRole, len(roles))
	for i, role := range roles {
		pbRoles[i] = &pb.UserRole{
			Id:        role.ID().String(),
			UserId:    role.UserID().String(),
			Role:      string(role.Role()),
			GrantedAt: timestamppb.New(role.GrantedAt()),
			IsActive:  role.IsActive(),
		}
	}

	return &pb.GetUserRolesResponse{
		Roles: pbRoles,
	}, nil
}

// Helper methods
func (h *AuthHandler) mapUserToPB(user *domain.User) *pb.User {
	return &pb.User{
		Id:          user.ID().String(),
		Email:       user.Email(),
		Username:    user.Username(),
		DisplayName: user.DisplayName(),
		IsVerified:  user.IsVerified(),
		IsActive:    user.IsActive(),
		CreatedAt:   timestamppb.New(user.CreatedAt()),
		UpdatedAt:   timestamppb.New(user.UpdatedAt()),
	}
}

func (h *AuthHandler) handleServiceError(err error) error {
	switch err.Error() {
	case "user not found":
		return status.Errorf(codes.NotFound, "user not found")
	case "invalid email or password":
		return status.Errorf(codes.Unauthenticated, "invalid email or password")
	case "user account is inactive":
		return status.Errorf(codes.PermissionDenied, "user account is inactive")
	case "email is already verified":
		return status.Errorf(codes.AlreadyExists, "email is already verified")
	case "user with this email already exists":
		return status.Errorf(codes.AlreadyExists, "user with this email already exists")
	case "user with this username already exists":
		return status.Errorf(codes.AlreadyExists, "user with this username already exists")
	case "invalid or expired refresh token":
		return status.Errorf(codes.Unauthenticated, "invalid or expired refresh token")
	case "current password is incorrect":
		return status.Errorf(codes.InvalidArgument, "current password is incorrect")
	case "email verification not found":
		return status.Errorf(codes.NotFound, "email verification token not found")
	case "email verification token has expired":
		return status.Errorf(codes.InvalidArgument, "email verification token has expired")
	case "email verification token has already been used":
		return status.Errorf(codes.InvalidArgument, "email verification token has already been used")
	case "email verification token is invalid":
		return status.Errorf(codes.InvalidArgument, "email verification token is invalid")
	default:
		h.logger.Error("Unhandled service error", logger.Error(err))
		return status.Errorf(codes.Internal, "internal server error")
	}
}
