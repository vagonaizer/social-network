package domain

import "social-network/user-service/internal/domain"

type UserBlacklistRepository interface {
	Save(blacklist *domain.UserBlacklist) error
	FindByUsers(userID, blockedUserID int64) (*domain.UserBlacklist, error)
	FindBlockedUsersByUserID(userID int64) ([]*domain.UserBlacklist, error)
	Delete(id int64) error
	IsBlocked(userID, blockedUserID int64) (bool, error)
}
