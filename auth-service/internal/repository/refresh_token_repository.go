package repository

import (
	"social-network/auth-service/internal/domain"

	"github.com/google/uuid"
)

type RefreshTokenRepository interface {
	Create(token *domain.RefreshToken) error
	GetByToken(token string) (*domain.RefreshToken, error)
	GetByUserID(userID uuid.UUID) ([]*domain.RefreshToken, error)
	Update(token *domain.RefreshToken) error
	Delete(id uuid.UUID) error
	DeleteByUserID(userID uuid.UUID) error
}
