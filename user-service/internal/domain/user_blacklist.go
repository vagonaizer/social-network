package domain

import (
	"errors"
	"time"
)

// UserBlacklist - запись в черном списке
type UserBlacklist struct {
	id            int64
	userID        int64
	blockedUserID int64
	blockedAt     time.Time
	reason        *string // опциональная причина блокировки
}

// Конструктор
func NewUserBlacklist(userID, blockedUserID int64, reason *string) (*UserBlacklist, error) {
	if userID == blockedUserID {
		return nil, errors.New("user cannot block themselves")
	}
	if userID <= 0 || blockedUserID <= 0 {
		return nil, errors.New("invalid user IDs")
	}

	return &UserBlacklist{
		userID:        userID,
		blockedUserID: blockedUserID,
		blockedAt:     time.Now(),
		reason:        reason,
	}, nil
}

// Геттеры
func (ub *UserBlacklist) ID() int64            { return ub.id }
func (ub *UserBlacklist) UserID() int64        { return ub.userID }
func (ub *UserBlacklist) BlockedUserID() int64 { return ub.blockedUserID }
func (ub *UserBlacklist) BlockedAt() time.Time { return ub.blockedAt }
func (ub *UserBlacklist) Reason() *string      { return ub.reason }

// Сеттеры
func (ub *UserBlacklist) SetReason(reason *string) {
	ub.reason = reason
}
