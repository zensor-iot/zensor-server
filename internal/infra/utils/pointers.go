package utils

import "time"

// StringPtr returns a pointer to the string if it's not empty, otherwise returns nil
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// TimePtr returns a pointer to the time if it's not zero, otherwise returns nil
func TimePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
