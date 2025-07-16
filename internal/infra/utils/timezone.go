package utils

import (
	"fmt"
	"time"
)

// ValidateTimezone validates that the given timezone string is a valid IANA timezone name
func ValidateTimezone(timezone string) error {
	if timezone == "" {
		return fmt.Errorf("timezone cannot be empty")
	}

	// Try to load the location to validate it's a valid IANA timezone
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone '%s': %w", timezone, err)
	}

	return nil
}

// IsValidTimezone checks if the given timezone string is a valid IANA timezone name
func IsValidTimezone(timezone string) bool {
	return ValidateTimezone(timezone) == nil
}
