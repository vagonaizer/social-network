package domain

import (
	"time"

	"github.com/google/uuid"
)

type PasswordReset struct {
	id        uuid.UUID
	userID    uuid.UUID
	token     string
	expiresAt time.Time
	isUsed    bool
	createdAt time.Time
}

// Constructor
func NewPasswordReset(userID uuid.UUID, token string, expiresAt time.Time) *PasswordReset {
	return &PasswordReset{
		id:        uuid.New(),
		userID:    userID,
		token:     token,
		expiresAt: expiresAt,
		isUsed:    false,
		createdAt: time.Now(),
	}
}

// Getters
func (pr *PasswordReset) ID() uuid.UUID {
	return pr.id
}

func (pr *PasswordReset) UserID() uuid.UUID {
	return pr.userID
}

func (pr *PasswordReset) Token() string {
	return pr.token
}

func (pr *PasswordReset) ExpiresAt() time.Time {
	return pr.expiresAt
}

func (pr *PasswordReset) IsUsed() bool {
	return pr.isUsed
}

func (pr *PasswordReset) CreatedAt() time.Time {
	return pr.createdAt
}

// Setters
func (pr *PasswordReset) SetUsed(used bool) {
	pr.isUsed = used
}

func (pr *PasswordReset) SetID(id uuid.UUID) {
	pr.id = id
}

func (pr *PasswordReset) SetUserID(userID uuid.UUID) {
	pr.userID = userID
}

func (pr *PasswordReset) SetToken(token string) {
	pr.token = token
}

func (pr *PasswordReset) SetExpiresAt(expiresAt time.Time) {
	pr.expiresAt = expiresAt
}

func (pr *PasswordReset) SetCreatedAt(createdAt time.Time) {
	pr.createdAt = createdAt
}

// Business methods
func (pr *PasswordReset) IsExpired() bool {
	return time.Now().After(pr.expiresAt)
}

func (pr *PasswordReset) IsValid() bool {
	return !pr.isUsed && !pr.IsExpired()
}
