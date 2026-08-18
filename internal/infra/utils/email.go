package utils

import (
	"errors"
	"fmt"
	"regexp"
)

var (
	errEmailCannotBeEmpty = errors.New("email cannot be empty")
	errInvalidEmailFormat = errors.New("invalid email format")
)

// ValidateEmail validates that the given email string is a valid email address.
func ValidateEmail(email string) error {
	if email == "" {
		return errEmailCannotBeEmpty
	}

	// Basic email regex pattern
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format '%s': %w", email, errInvalidEmailFormat)
	}

	return nil
}

// IsValidEmail checks if the given email string is a valid email address.
func IsValidEmail(email string) bool {
	return ValidateEmail(email) == nil
}
