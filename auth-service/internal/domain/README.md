### Auth Service Domain Models

## Overview

This document describes the domain models for the authentication service. All models follow Domain-Driven Design principles with proper encapsulation and business logic separation.

## Core Principles

# Encapsulation

All struct fields are unexported (lowercase) to prevent direct access from outside the package. Access is provided through getter and setter methods, ensuring data integrity and allowing for future validation or business logic injection.

```
type UserAuth struct {
	id           uuid.UUID
	userID       uuid.UUID
	passwordHash string
	lastLoginAt  *time.Time
	createdAt    time.Time
	updatedAt    time.Time
}
```

instead of: 

```
type UserAuth struct {
	Id           uuid.UUID
	UserID       uuid.UUID
	PasswordHash string
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
```

# UUID Usage

All entities use UUID as primary identifiers instead of auto-incrementing integers. This provides better distribution in microservices architecture and eliminates ID collision risks.

I don't want to write any silly ID-GENERATION methods, because it's probably just gonna suck.

# Immutable Creation

Entities are created through constructor functions that set default values and generate UUIDs automatically:

```
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
```

### Domain Entities

## User

Represents the core user entity in the system.

```
type User struct {
    id          uuid.UUID  // Unique identifier
    email       string     // User's email address (unique)
    username    string     // User's username (unique)
    displayName string     // Human-readable display name
    isVerified  bool       // Email verification status
    isActive    bool       // Account active status
    createdAt   time.Time  // Account creation timestamp
    updatedAt   time.Time  // Last modification timestamp
}
```
# Key Points: 

`isVerified` controls whether user can login
`isActive` allows for soft account deactivation
`updatedAt` is automatically set by setters


## UserAuth

Contains authentication-specific data separated from user profile.

```
type UserAuth struct {
    id           uuid.UUID   // Unique identifier
    userID       uuid.UUID   // Reference to User entity
    passwordHash string      // Bcrypt hashed password
    lastLoginAt  *time.Time  // Last successful login (nullable)
    createdAt    time.Time   // Creation timestamp
    updatedAt    time.Time   // Last modification timestamp
}
```

# Key Points:

Separated from User for security and performance
`lastLoginAt` is pointer to allow null values
Password is always stored as hash, never plaintext

## UserRole

Defines user roles with type safety and extensibility.

```
type UserRoleType string

const (
    RoleAdmin     UserRoleType = "admin"
    RoleModerator UserRoleType = "moderator"
    RoleUser      UserRoleType = "user"
)

type UserRole struct {
    id        uuid.UUID     // Unique identifier
    userID    uuid.UUID     // Reference to User entity
    role      UserRoleType  // Role type (type-safe enum)
    grantedAt time.Time     // When role was granted
    isActive  bool          // Role active status
}
```

**Key Points:**

- Uses type alias for role values to prevent typos
- Supports role deactivation without deletion
- Can be extended with additional fields (grantedBy, expiresAt)

**Business Methods:**

- `IsAdmin()`, `IsModerator()`, `IsUser()` - Check active role status

### RefreshToken

Manages JWT refresh tokens for session management.

```
type RefreshToken struct {
    id        uuid.UUID  // Unique identifier
    userID    uuid.UUID  // Reference to User entity
    token     string     // Actual refresh token value
    expiresAt time.Time  // Token expiration time
    isRevoked bool       // Revocation status
    createdAt time.Time  // Creation timestamp
}
```

**Key Points:**

- Allows token revocation without waiting for expiration
- Supports multiple active tokens per user
- Immutable after creation (except revocation)


**Business Methods:**

- `IsExpired()` - Check if token has expired
- `IsValid()` - Check if token is both active and not expired

### EmailVerification

Handles email verification process.

```
type EmailVerification struct {
    id        uuid.UUID  // Unique identifier
    userID    uuid.UUID  // Reference to User entity
    token     string     // Verification token
    expiresAt time.Time  // Token expiration time
    isUsed    bool       // Usage status
    createdAt time.Time  // Creation timestamp
}
```

**Key Points:**

- One-time use tokens
- Time-limited validity
- Prevents replay attacks


**Business Methods:**

- `IsExpired()` - Check token expiration
- `IsValid()` - Check if token is unused and not expired

### PasswordReset

Manages password reset functionality.

```
type PasswordReset struct {
    id        uuid.UUID  // Unique identifier
    userID    uuid.UUID  // Reference to User entity
    token     string     // Reset token
    expiresAt time.Time  // Token expiration time
    isUsed    bool       // Usage status
    createdAt time.Time  // Creation timestamp
}
```

**Key Points:**

- Similar to EmailVerification but for password resets
- Short expiration time for security
- Single-use tokens


**Business Methods:**

- `IsExpired()` - Check token expiration
- `IsValid()` - Check if token is unused and not expired

## Benefits of This Design

**Type Safety:** Role constants prevent typos and invalid values.

**Encapsulation:** Private fields ensure controlled access and data integrity.

**Extensibility:** Struct-based design allows adding fields without breaking existing code.

**Business Logic:** Domain methods encapsulate business rules and validation.

**Separation of Concerns:** Each entity has a single responsibility.

**Testability:** Pure domain logic without external dependencies.