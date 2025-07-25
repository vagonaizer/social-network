package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserAuth struct {
	id           uuid.UUID
	userID       uuid.UUID
	passwordHash string
	lastLoginAt  *time.Time
	createdAt    time.Time
	updatedAt    time.Time
}

// Constructor
func NewUserAuth(userID uuid.UUID, passwordHash string) *UserAuth {
	return &UserAuth{
		id:           uuid.New(),
		userID:       userID,
		passwordHash: passwordHash,
		lastLoginAt:  nil,
		createdAt:    time.Now(),
		updatedAt:    time.Now(),
	}
}

// Getters
func (ua *UserAuth) ID() uuid.UUID {
	return ua.id
}

func (ua *UserAuth) UserID() uuid.UUID {
	return ua.userID
}

func (ua *UserAuth) PasswordHash() string {
	return ua.passwordHash
}

func (ua *UserAuth) LastLoginAt() *time.Time {
	return ua.lastLoginAt
}

func (ua *UserAuth) CreatedAt() time.Time {
	return ua.createdAt
}

func (ua *UserAuth) UpdatedAt() time.Time {
	return ua.updatedAt
}

// Setters
func (ua *UserAuth) SetPasswordHash(passwordHash string) {
	ua.passwordHash = passwordHash
	ua.updatedAt = time.Now()
}

func (ua *UserAuth) SetLastLoginAt(lastLoginAt *time.Time) {
	ua.lastLoginAt = lastLoginAt
	ua.updatedAt = time.Now()
}

func (ua *UserAuth) SetID(id uuid.UUID) {
	ua.id = id
}

func (ua *UserAuth) SetUserID(userID uuid.UUID) {
	ua.userID = userID
}

func (ua *UserAuth) SetCreatedAt(createdAt time.Time) {
	ua.createdAt = createdAt
}

func (ua *UserAuth) SetUpdatedAt(updatedAt time.Time) {
	ua.updatedAt = updatedAt
}
