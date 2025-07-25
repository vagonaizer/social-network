package repository

import (
	"social-network/auth-service/internal/domain"

	"github.com/google/uuid"
)

type EmailVerificationRepository interface {
	Create(verification *domain.EmailVerification) error
	GetByToken(token string) (*domain.EmailVerification, error)
	GetByUserID(userID uuid.UUID) (*domain.EmailVerification, error)
	Update(verification *domain.EmailVerification) error
	Delete(id uuid.UUID) error
}
