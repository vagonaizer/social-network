package repository

import "errors"

// Ошибки репозиториев
var (
	ErrUserNotFound           = errors.New("user not found")
	ErrFriendshipNotFound     = errors.New("friendship not found")
	ErrBlacklistEntryNotFound = errors.New("blacklist entry not found")
	ErrDuplicateEntry         = errors.New("duplicate entry")
)
