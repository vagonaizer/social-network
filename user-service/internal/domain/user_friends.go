package domain

import (
	"errors"
	"time"
)

// FriendshipStatus - статус дружбы
type FriendshipStatus int

const (
	FriendshipPending FriendshipStatus = iota
	FriendshipAccepted
	FriendshipRejected
)

// Friendship - отношения дружбы между пользователями
type Friendship struct {
	id          int64
	requesterID int64
	addresseeID int64
	status      FriendshipStatus
	createdAt   time.Time
	updatedAt   time.Time
}

// Конструктор
func NewFriendship(requesterID, addresseeID int64) (*Friendship, error) {
	if requesterID == addresseeID {
		return nil, errors.New("user cannot be friend with themselves")
	}
	if requesterID <= 0 || addresseeID <= 0 {
		return nil, errors.New("invalid user IDs")
	}

	now := time.Now()
	return &Friendship{
		requesterID: requesterID,
		addresseeID: addresseeID,
		status:      FriendshipPending,
		createdAt:   now,
		updatedAt:   now,
	}, nil
}

// Геттеры
func (f *Friendship) ID() int64                { return f.id }
func (f *Friendship) RequesterID() int64       { return f.requesterID }
func (f *Friendship) AddresseeID() int64       { return f.addresseeID }
func (f *Friendship) Status() FriendshipStatus { return f.status }
func (f *Friendship) CreatedAt() time.Time     { return f.createdAt }
func (f *Friendship) UpdatedAt() time.Time     { return f.updatedAt }

// Бизнес-методы
func (f *Friendship) Accept() error {
	if f.status != FriendshipPending {
		return errors.New("can only accept pending friendship")
	}
	f.status = FriendshipAccepted
	f.updatedAt = time.Now()
	return nil
}

func (f *Friendship) Reject() error {
	if f.status != FriendshipPending {
		return errors.New("can only reject pending friendship")
	}
	f.status = FriendshipRejected
	f.updatedAt = time.Now()
	return nil
}

func (f *Friendship) IsAccepted() bool {
	return f.status == FriendshipAccepted
}

func (f *Friendship) IsPending() bool {
	return f.status == FriendshipPending
}

func (f *Friendship) GetOtherUserID(userID int64) (int64, error) {
	if userID == f.requesterID {
		return f.addresseeID, nil
	}
	if userID == f.addresseeID {
		return f.requesterID, nil
	}
	return 0, errors.New("user is not part of this friendship")
}
