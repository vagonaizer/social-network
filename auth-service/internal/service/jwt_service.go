package service

import (
	"social-network/auth-service/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	accessSecret  []byte
	refreshSecret []byte
	issuer        string
}

type AccessTokenClaims struct {
	UserID      uuid.UUID             `json:"user_id"`
	Email       string                `json:"email"`
	Username    string                `json:"username"`
	DisplayName string                `json:"display_name"`
	Roles       []domain.UserRoleType `json:"roles"`
	IsVerified  bool                  `json:"is_verified"`
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func NewJWTService(accessSecret, refreshSecret []byte, issuer string) *JWTService {
	return &JWTService{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		issuer:        issuer,
	}
}

// GenerateAccessToken создает access token
func (s *JWTService) GenerateAccessToken(user *domain.User, roles []domain.UserRoleType) (string, error) {
	now := time.Now()
	claims := AccessTokenClaims{
		UserID:      user.ID(),
		Email:       user.Email(),
		Username:    user.Username(),
		DisplayName: user.DisplayName(),
		Roles:       roles,
		IsVerified:  user.IsVerified(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   user.ID().String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)), // 15 минут
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.accessSecret)
}

// GenerateRefreshToken создает refresh token
func (s *JWTService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := RefreshTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)), // 7 дней
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.refreshSecret)
}

// ValidateAccessToken проверяет access token
func (s *JWTService) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return s.accessSecret, nil
	})

	if err != nil {
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// ValidateRefreshToken проверяет refresh token
func (s *JWTService) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return s.refreshSecret, nil
	})

	if err != nil {
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// ExtractUserIDFromToken извлекает ID пользователя из токена без полной валидации
func (s *JWTService) ExtractUserIDFromToken(tokenString string) (uuid.UUID, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &AccessTokenClaims{})
	if err != nil {
		return uuid.Nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(*AccessTokenClaims); ok {
		return claims.UserID, nil
	}

	return uuid.Nil, ErrTokenInvalid
}
