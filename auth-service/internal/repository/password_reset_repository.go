package repository

import (
	"social-network/auth-service/internal/domain"

	"github.com/google/uuid"
)

type PasswordResetRepository interface {
	Create(reset *domain.PasswordReset) error
	GetByToken(token string) (*domain.PasswordReset, error)
	GetByUserID(userID uuid.UUID) (*domain.PasswordReset, error)
	Update(reset *domain.PasswordReset) error
	Delete(id uuid.UUID) error
}
