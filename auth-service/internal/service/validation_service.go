package service

import "social-network/auth-service/pkg/helpers"

type ValidationService struct{}

func NewValidationService() *ValidationService {
	return &ValidationService{}
}

// ValidateEmail проверяет формат email
func (s *ValidationService) ValidateEmail(email string) error {
	if !helpers.ValidateEmail(email) {
		return ErrInvalidEmailFormat
	}
	return nil
}

// ValidateUsername проверяет формат username
func (s *ValidationService) ValidateUsername(username string) error {
	if !helpers.ValidateUsername(username) {
		return ErrInvalidUsernameFormat
	}
	return nil
}

// ValidatePassword проверяет надежность пароля
func (s *ValidationService) ValidatePassword(password string) error {
	if !helpers.ValidatePassword(password) {
		return ErrPasswordTooWeak
	}
	return nil
}

// ValidateDisplayName проверяет отображаемое имя
func (s *ValidationService) ValidateDisplayName(displayName string) error {
	if !helpers.ValidateDisplayName(displayName) {
		return ErrInvalidDisplayName
	}
	return nil
}

// ValidateRegistrationData проверяет все данные регистрации
func (s *ValidationService) ValidateRegistrationData(email, username, displayName, password string) error {
	if err := s.ValidateEmail(email); err != nil {
		return err
	}

	if err := s.ValidateUsername(username); err != nil {
		return err
	}

	if err := s.ValidateDisplayName(displayName); err != nil {
		return err
	}

	if err := s.ValidatePassword(password); err != nil {
		return err
	}

	return nil
}
