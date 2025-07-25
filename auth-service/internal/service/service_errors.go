package service

import "errors"

// Authentication Errors
var (
	// ErrInvalidCredentials is returned when email/password combination is invalid
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrUserInactive is returned when user account is deactivated
	ErrUserInactive = errors.New("user account is inactive")

	// ErrEmailAlreadyVerified is returned when trying to verify already verified email
	ErrEmailAlreadyVerified = errors.New("email is already verified")

	// ErrInvalidCurrentPassword is returned when current password is incorrect during password change
	ErrInvalidCurrentPassword = errors.New("current password is incorrect")
)

// Token Errors
var (
	// ErrTokenExpired is returned when token has expired
	ErrTokenExpired = errors.New("token has expired")

	// ErrTokenInvalid is returned when token is invalid or malformed
	ErrTokenInvalid = errors.New("token is invalid")

	// ErrTokenRevoked is returned when token has been revoked
	ErrTokenRevoked = errors.New("token has been revoked")
)

// Permission Errors
var (
	// ErrInsufficientPermissions is returned when user doesn't have required permissions
	ErrInsufficientPermissions = errors.New("insufficient permissions")

	// ErrUnauthorized is returned when user is not authenticated
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when user is authenticated but access is forbidden
	ErrForbidden = errors.New("forbidden")
)

// Validation Errors
var (
	// ErrInvalidEmailFormat is returned when email format is invalid
	ErrInvalidEmailFormat = errors.New("invalid email format")

	// ErrInvalidUsernameFormat is returned when username format is invalid
	ErrInvalidUsernameFormat = errors.New("invalid username format")

	// ErrPasswordTooWeak is returned when password doesn't meet security requirements
	ErrPasswordTooWeak = errors.New("password is too weak")

	// ErrInvalidDisplayName is returned when display name is invalid
	ErrInvalidDisplayName = errors.New("invalid display name")
)

// Rate Limiting Errors
var (
	// ErrRateLimitExceeded is returned when rate limit is exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")

	// ErrTooManyAttempts is returned when too many failed attempts are made
	ErrTooManyAttempts = errors.New("too many failed attempts")
)
