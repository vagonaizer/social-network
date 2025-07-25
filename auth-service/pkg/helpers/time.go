package helpers

import "time"

// TokenExpirationTimes содержит стандартные времена истечения токенов
var TokenExpirationTimes = struct {
	AccessToken       time.Duration
	RefreshToken      time.Duration
	EmailVerification time.Duration
	PasswordReset     time.Duration
}{
	AccessToken:       15 * time.Minute,
	RefreshToken:      7 * 24 * time.Hour,
	EmailVerification: 24 * time.Hour,
	PasswordReset:     1 * time.Hour,
}

// GetExpirationTime возвращает время истечения для токена
func GetExpirationTime(tokenType string) time.Time {
	now := time.Now()

	switch tokenType {
	case "access":
		return now.Add(TokenExpirationTimes.AccessToken)
	case "refresh":
		return now.Add(TokenExpirationTimes.RefreshToken)
	case "email_verification":
		return now.Add(TokenExpirationTimes.EmailVerification)
	case "password_reset":
		return now.Add(TokenExpirationTimes.PasswordReset)
	default:
		return now.Add(1 * time.Hour) // default 1 hour
	}
}

// IsExpired проверяет, истек ли токен
func IsExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}

// TimeUntilExpiration возвращает время до истечения
func TimeUntilExpiration(expiresAt time.Time) time.Duration {
	return time.Until(expiresAt)
}
