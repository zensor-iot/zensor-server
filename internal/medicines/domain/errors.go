package domain

import "errors"

var (
	ErrTenantIDRequired       = errors.New("tenant id is required")
	ErrPatientIDRequired      = errors.New("patient id is required")
	ErrPatientNameRequired    = errors.New("patient name is required")
	ErrInvalidPatientKind     = errors.New("patient kind is invalid")
	ErrMedicineNameRequired   = errors.New("medicine name is required")
	ErrDoseQuantityRequired   = errors.New("dose quantity must be greater than zero")
	ErrInvalidDoseUnit        = errors.New("dose unit is invalid")
	ErrInvalidDoseStatus      = errors.New("dose status is invalid")
	ErrTreatmentIDRequired    = errors.New("treatment id is required")
	ErrScheduledAtRequired    = errors.New("scheduled date is required")
	ErrSequenceNumberInvalid  = errors.New("dose sequence number must not be negative")
	ErrStartAtRequired        = errors.New("schedule start date is required")
	ErrIntervalRequired       = errors.New("schedule interval must be greater than zero")
	ErrInvalidIntervalUnit    = errors.New("schedule interval unit is invalid")
	ErrScheduleWindowTooLarge = errors.New("schedule window produces too many occurrences")
	ErrTotalDosesInvalid      = errors.New("total doses must be greater than zero")
	ErrEndAtBeforeStartAt     = errors.New("treatment end date must be after the schedule start date")
	ErrDoseAlreadyResolved    = errors.New("dose is already administered or skipped")
)
