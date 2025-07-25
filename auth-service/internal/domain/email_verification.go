package domain

import (
	"time"

	"github.com/google/uuid"
)

type EmailVerification struct {
	id        uuid.UUID
	userID    uuid.UUID
	token     string
	expiresAt time.Time
	isUsed    bool
	createdAt time.Time
}

// Constructor
func NewEmailVerification(userID uuid.UUID, token string, expiresAt time.Time) *EmailVerification {
	return &EmailVerification{
		id:        uuid.New(),
		userID:    userID,
		token:     token,
		expiresAt: expiresAt,
		isUsed:    false,
		createdAt: time.Now(),
	}
}

// Getters
func (ev *EmailVerification) ID() uuid.UUID {
	return ev.id
}

func (ev *EmailVerification) UserID() uuid.UUID {
	return ev.userID
}

func (ev *EmailVerification) Token() string {
	return ev.token
}

func (ev *EmailVerification) ExpiresAt() time.Time {
	return ev.expiresAt
}

func (ev *EmailVerification) IsUsed() bool {
	return ev.isUsed
}

func (ev *EmailVerification) CreatedAt() time.Time {
	return ev.createdAt
}

// Setters
func (ev *EmailVerification) SetUsed(used bool) {
	ev.isUsed = used
}

func (ev *EmailVerification) SetID(id uuid.UUID) {
	ev.id = id
}

func (ev *EmailVerification) SetUserID(userID uuid.UUID) {
	ev.userID = userID
}

func (ev *EmailVerification) SetToken(token string) {
	ev.token = token
}

func (ev *EmailVerification) SetExpiresAt(expiresAt time.Time) {
	ev.expiresAt = expiresAt
}

func (ev *EmailVerification) SetCreatedAt(createdAt time.Time) {
	ev.createdAt = createdAt
}

// Business methods
func (ev *EmailVerification) IsExpired() bool {
	return time.Now().After(ev.expiresAt)
}

func (ev *EmailVerification) IsValid() bool {
	return !ev.isUsed && !ev.IsExpired()
}
