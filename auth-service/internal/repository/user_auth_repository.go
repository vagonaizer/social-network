package repository

import (
	"social-network/auth-service/internal/domain"

	"github.com/google/uuid"
)

type UserAuthRepository interface {
	Create(userAuth *domain.UserAuth) error
	GetByUserID(userID uuid.UUID) (*domain.UserAuth, error)
	Update(userAuth *domain.UserAuth) error
	Delete(userID uuid.UUID) error
}
