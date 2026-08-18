package domain

import "errors"

var (
	ErrActivityIDRequired    = errors.New("activity ID is required")
	ErrScheduledDateRequired = errors.New("scheduled date is required")
	ErrStartDateRequired     = errors.New("schedule start date is required")
	ErrIntervalRequired      = errors.New("schedule interval must be greater than zero")
	ErrInvalidRecurrenceUnit = errors.New("schedule recurrence unit is invalid")
	ErrScheduleRequired      = errors.New("schedule is required")
)
