package repository

import (
	"social-network/auth-service/internal/domain"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(user *domain.User) error
	GetByID(id uuid.UUID) (*domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	GetByUsername(username string) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id uuid.UUID) error

	ExistsByEmail(email string) (bool, error)
	ExistsByUsername(username string) (bool, error)
}
