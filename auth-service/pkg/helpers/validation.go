package helpers

import (
	"regexp"
	"strings"
	"unicode"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)
	numericRegex  = regexp.MustCompile(`^\d+$`)
)

// ValidateEmail проверяет формат email
func ValidateEmail(email string) bool {
	if email == "" || len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

// ValidateUsername проверяет формат username
func ValidateUsername(username string) bool {
	if username == "" || len(username) < 3 || len(username) > 30 {
		return false
	}

	if !usernameRegex.MatchString(username) {
		return false
	}

	// Проверяем, что username не состоит только из цифр
	return !numericRegex.MatchString(username)
}

// ValidatePassword проверяет надежность пароля
func ValidatePassword(password string) bool {
	if len(password) < 8 || len(password) > 128 {
		return false
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// ValidateDisplayName проверяет отображаемое имя
func ValidateDisplayName(displayName string) bool {
	if displayName == "" {
		return false
	}

	displayName = strings.TrimSpace(displayName)
	if len(displayName) < 1 || len(displayName) > 100 {
		return false
	}

	// Проверяем, что имя не состоит только из пробелов
	return strings.TrimSpace(displayName) != ""
}

// SanitizeString очищает строку от лишних пробелов
func SanitizeString(s string) string {
	return strings.TrimSpace(s)
}

// IsValidUUID проверяет формат UUID (простая проверка)
func IsValidUUID(uuid string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(uuid)
}
