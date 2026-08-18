package utils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	errTimezoneCannotBeEmpty = errors.New("timezone cannot be empty")
	errExecutionTimeFormat   = errors.New("execution time must be in HH:MM format")
	errHourOutOfRange        = errors.New("hour must be between 0 and 23")
	errMinuteOutOfRange      = errors.New("minute must be between 0 and 59")
)

// ValidateTimezone validates that the given timezone string is a valid IANA timezone name.
func ValidateTimezone(timezone string) error {
	if timezone == "" {
		return errTimezoneCannotBeEmpty
	}

	// Try to load the location to validate it's a valid IANA timezone
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone '%s': %w", timezone, err)
	}

	return nil
}

// IsValidTimezone checks if the given timezone string is a valid IANA timezone name.
func IsValidTimezone(timezone string) bool {
	return ValidateTimezone(timezone) == nil
}

// ParseExecutionTime parses a time string like "02:00" or "14:30".
func ParseExecutionTime(timeStr string) (hour, minute int, err error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, 0, errExecutionTimeFormat
	}

	hour, err = strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, errHourOutOfRange
	}

	minute, err = strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, errMinuteOutOfRange
	}

	return hour, minute, nil
}
