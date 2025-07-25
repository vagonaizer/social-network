package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserRoleType string

const (
	RoleAdmin     UserRoleType = "admin"
	RoleModerator UserRoleType = "moderator"
	RoleUser      UserRoleType = "user"
)

// Instead of using only string value (alias) for roles we define a struct
// which can be extended in the future with more fields if needed.
type UserRole struct {
	id        uuid.UUID
	userID    uuid.UUID
	role      UserRoleType
	grantedAt time.Time
	isActive  bool
}

// Constructor
func NewUserRole(userID uuid.UUID, role UserRoleType) *UserRole {
	return &UserRole{
		id:        uuid.New(),
		userID:    userID,
		role:      role,
		grantedAt: time.Now(),
		isActive:  true,
	}
}

// Getters
func (ur *UserRole) ID() uuid.UUID {
	return ur.id
}

func (ur *UserRole) UserID() uuid.UUID {
	return ur.userID
}

func (ur *UserRole) Role() UserRoleType {
	return ur.role
}

func (ur *UserRole) GrantedAt() time.Time {
	return ur.grantedAt
}

func (ur *UserRole) IsActive() bool {
	return ur.isActive
}

// Setters
func (ur *UserRole) SetRole(role UserRoleType) {
	ur.role = role
}

func (ur *UserRole) SetActive(active bool) {
	ur.isActive = active
}

func (ur *UserRole) SetID(id uuid.UUID) {
	ur.id = id
}

func (ur *UserRole) SetUserID(userID uuid.UUID) {
	ur.userID = userID
}

func (ur *UserRole) SetGrantedAt(grantedAt time.Time) {
	ur.grantedAt = grantedAt
}

// Business methods
func (ur *UserRole) IsAdmin() bool {
	return ur.role == RoleAdmin && ur.isActive
}

func (ur *UserRole) IsModerator() bool {
	return ur.role == RoleModerator && ur.isActive
}

func (ur *UserRole) IsUser() bool {
	return ur.role == RoleUser && ur.isActive
}
