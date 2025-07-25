package domain

import "social-network/user-service/internal/domain"

type FriendshipRepository interface {
	Save(friendship *domain.Friendship) error
	FindByUsers(userID1, userID2 int64) (*domain.Friendship, error)
	FindFriendsByUserID(userID int64) ([]*domain.Friendship, error)
	FindPendingRequestsByUserID(userID int64) ([]*domain.Friendship, error)
	Update(friendship *domain.Friendship) error
	Delete(id int64) error
}
