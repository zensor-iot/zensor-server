package domain

import "errors"

var (
	ErrActivityIDRequired    = errors.New("activity ID is required")
	ErrScheduledDateRequired = errors.New("scheduled date is required")
)
