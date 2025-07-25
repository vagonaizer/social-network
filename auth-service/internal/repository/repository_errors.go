package repository

import "errors"

// User Repository Errors
var (
	// ErrUserNotFound is returned when a user cannot be found by the given criteria
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists is returned when trying to create a user that already exists
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrUserEmailExists is returned when trying to create a user with an email that already exists
	ErrUserEmailExists = errors.New("user with this email already exists")

	// ErrUserUsernameExists is returned when trying to create a user with a username that already exists
	ErrUserUsernameExists = errors.New("user with this username already exists")
)

// User Auth Repository Errors
var (
	// ErrUserAuthNotFound is returned when user authentication data cannot be found
	ErrUserAuthNotFound = errors.New("user auth not found")

	// ErrUserAuthAlreadyExists is returned when trying to create auth data for a user that already has it
	ErrUserAuthAlreadyExists = errors.New("user auth already exists")
)

// User Role Repository Errors
var (
	// ErrUserRoleNotFound is returned when a user role cannot be found
	ErrUserRoleNotFound = errors.New("user role not found")

	// ErrUserRoleAlreadyExists is returned when trying to assign a role that user already has
	ErrUserRoleAlreadyExists = errors.New("user role already exists")

	// ErrInvalidRole is returned when an invalid role type is provided
	ErrInvalidRole = errors.New("invalid role type")
)

// Refresh Token Repository Errors
var (
	// ErrRefreshTokenNotFound is returned when a refresh token cannot be found
	ErrRefreshTokenNotFound = errors.New("refresh token not found")

	// ErrRefreshTokenExpired is returned when a refresh token has expired
	ErrRefreshTokenExpired = errors.New("refresh token has expired")

	// ErrRefreshTokenRevoked is returned when a refresh token has been revoked
	ErrRefreshTokenRevoked = errors.New("refresh token has been revoked")

	// ErrRefreshTokenInvalid is returned when a refresh token is invalid (expired or revoked)
	ErrRefreshTokenInvalid = errors.New("refresh token is invalid")
)

// Email Verification Repository Errors
var (
	// ErrEmailVerificationNotFound is returned when an email verification record cannot be found
	ErrEmailVerificationNotFound = errors.New("email verification not found")

	// ErrEmailVerificationExpired is returned when an email verification token has expired
	ErrEmailVerificationExpired = errors.New("email verification token has expired")

	// ErrEmailVerificationUsed is returned when an email verification token has already been used
	ErrEmailVerificationUsed = errors.New("email verification token has already been used")

	// ErrEmailVerificationInvalid is returned when an email verification token is invalid (expired or used)
	ErrEmailVerificationInvalid = errors.New("email verification token is invalid")
)

// Password Reset Repository Errors
var (
	// ErrPasswordResetNotFound is returned when a password reset record cannot be found
	ErrPasswordResetNotFound = errors.New("password reset not found")

	// ErrPasswordResetExpired is returned when a password reset token has expired
	ErrPasswordResetExpired = errors.New("password reset token has expired")

	// ErrPasswordResetUsed is returned when a password reset token has already been used
	ErrPasswordResetUsed = errors.New("password reset token has already been used")

	// ErrPasswordResetInvalid is returned when a password reset token is invalid (expired or used)
	ErrPasswordResetInvalid = errors.New("password reset token is invalid")
)

// Database Connection Errors
var (
	// ErrDatabaseConnection is returned when there's a problem connecting to the database
	ErrDatabaseConnection = errors.New("database connection error")

	// ErrDatabaseTransaction is returned when there's a problem with database transaction
	ErrDatabaseTransaction = errors.New("database transaction error")

	// ErrDatabaseConstraint is returned when a database constraint is violated
	ErrDatabaseConstraint = errors.New("database constraint violation")

	// ErrDatabaseTimeout is returned when a database operation times out
	ErrDatabaseTimeout = errors.New("database operation timeout")
)

// Validation Errors
var (
	// ErrInvalidUUID is returned when an invalid UUID is provided
	ErrInvalidUUID = errors.New("invalid UUID format")

	// ErrInvalidEmail is returned when an invalid email format is provided
	ErrInvalidEmail = errors.New("invalid email format")

	// ErrInvalidUsername is returned when an invalid username is provided
	ErrInvalidUsername = errors.New("invalid username format")

	// ErrInvalidPassword is returned when an invalid password is provided
	ErrInvalidPassword = errors.New("invalid password format")

	// ErrPasswordTooShort is returned when password is shorter than minimum required length
	ErrPasswordTooShort = errors.New("password is too short")

	// ErrPasswordTooLong is returned when password is longer than maximum allowed length
	ErrPasswordTooLong = errors.New("password is too long")
)

// Generic Repository Errors
var (
	// ErrRecordNotFound is a generic error for when any record is not found
	ErrRecordNotFound = errors.New("record not found")

	// ErrRecordAlreadyExists is a generic error for when a record already exists
	ErrRecordAlreadyExists = errors.New("record already exists")

	// ErrInvalidInput is returned when invalid input is provided to repository methods
	ErrInvalidInput = errors.New("invalid input provided")

	// ErrOperationFailed is returned when a repository operation fails for unknown reasons
	ErrOperationFailed = errors.New("repository operation failed")
)
