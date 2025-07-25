package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	id          uuid.UUID
	email       string
	username    string
	displayName string
	isVerified  bool // Indicates if the user's email is verified
	isActive    bool // Indicates if the user account is active / may be suspended
	createdAt   time.Time
	updatedAt   time.Time
}

func NewUser(email, username, displayName string) *User {
	return &User{
		id:          uuid.New(),
		email:       email,
		username:    username,
		displayName: displayName,
		isVerified:  false,
		isActive:    true,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}
}

// Getters
func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Email() string {
	return u.email
}

func (u *User) Username() string {
	return u.username
}

func (u *User) DisplayName() string {
	return u.displayName
}

func (u *User) IsVerified() bool {
	return u.isVerified
}

func (u *User) IsActive() bool {
	return u.isActive
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}

func (u *User) UpdatedAt() time.Time {
	return u.updatedAt
}

// Setters
func (u *User) SetEmail(email string) {
	u.email = email
	u.updatedAt = time.Now()
}

func (u *User) SetUsername(username string) {
	u.username = username
	u.updatedAt = time.Now()
}

func (u *User) SetDisplayName(displayName string) {
	u.displayName = displayName
	u.updatedAt = time.Now()
}

func (u *User) SetVerified(verified bool) {
	u.isVerified = verified
	u.updatedAt = time.Now()
}

func (u *User) SetActive(active bool) {
	u.isActive = active
	u.updatedAt = time.Now()
}

func (u *User) SetID(id uuid.UUID) {
	u.id = id
}

func (u *User) SetCreatedAt(createdAt time.Time) {
	u.createdAt = createdAt
}

func (u *User) SetUpdatedAt(updatedAt time.Time) {
	u.updatedAt = updatedAt
}
