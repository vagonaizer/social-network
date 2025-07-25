package repository

import (
	"social-network/auth-service/internal/domain"

	"github.com/google/uuid"
)

type UserRoleRepository interface {
	Create(userRole *domain.UserRole) error
	GetByUserID(userID uuid.UUID) ([]*domain.UserRole, error)
	Update(userRole *domain.UserRole) error
	Delete(id uuid.UUID) error
}
