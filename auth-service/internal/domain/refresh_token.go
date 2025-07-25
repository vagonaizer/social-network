package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	id        uuid.UUID
	userID    uuid.UUID
	token     string
	expiresAt time.Time
	isRevoked bool
	createdAt time.Time
}

// Constructor
func NewRefreshToken(userID uuid.UUID, token string, expiresAt time.Time) *RefreshToken {
	return &RefreshToken{
		id:        uuid.New(),
		userID:    userID,
		token:     token,
		expiresAt: expiresAt,
		isRevoked: false,
		createdAt: time.Now(),
	}
}

// Getters
func (rt *RefreshToken) ID() uuid.UUID {
	return rt.id
}

func (rt *RefreshToken) UserID() uuid.UUID {
	return rt.userID
}

func (rt *RefreshToken) Token() string {
	return rt.token
}

func (rt *RefreshToken) ExpiresAt() time.Time {
	return rt.expiresAt
}

func (rt *RefreshToken) IsRevoked() bool {
	return rt.isRevoked
}

func (rt *RefreshToken) CreatedAt() time.Time {
	return rt.createdAt
}

// Setters
func (rt *RefreshToken) SetRevoked(revoked bool) {
	rt.isRevoked = revoked
}

func (rt *RefreshToken) SetID(id uuid.UUID) {
	rt.id = id
}

func (rt *RefreshToken) SetUserID(userID uuid.UUID) {
	rt.userID = userID
}

func (rt *RefreshToken) SetToken(token string) {
	rt.token = token
}

func (rt *RefreshToken) SetExpiresAt(expiresAt time.Time) {
	rt.expiresAt = expiresAt
}

func (rt *RefreshToken) SetCreatedAt(createdAt time.Time) {
	rt.createdAt = createdAt
}

// Business methods
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.expiresAt)
}

func (rt *RefreshToken) IsValid() bool {
	return !rt.isRevoked && !rt.IsExpired()
}
