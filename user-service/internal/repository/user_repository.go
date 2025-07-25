package repository

import "social-network/user-service/internal/domain"

type UserRepository interface {
	Save(user *domain.User) error
	FindByID(id int64) (*domain.User, error)
	FindByShortUrlName(shortUrlName string) (*domain.User, error)
	Update(user *domain.User) error
	Delete(id int64) error
}
