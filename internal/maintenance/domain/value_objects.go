package domain

import (
	"time"
)

type (
	CompletionNotes string
	CompletedBy     string
	CustomTypeName  string
	Days            []int
	OverdueDays     int
)

type RecurrenceUnit string

const (
	RecurrenceUnitDay     RecurrenceUnit = "day"
	RecurrenceUnitWeek    RecurrenceUnit = "week"
	RecurrenceUnitMonth   RecurrenceUnit = "month"
	RecurrenceUnitQuarter RecurrenceUnit = "quarter"
	RecurrenceUnitYear    RecurrenceUnit = "year"
)

type Schedule struct {
	StartDate time.Time      `json:"start_date"`
	Every     int            `json:"every"`
	Unit      RecurrenceUnit `json:"unit"`
}

func (s Schedule) Next(after time.Time) (time.Time, error) {
	if err := s.validate(); err != nil {
		return time.Time{}, err
	}

	next := s.StartDate
	for next.Before(after) || next.Equal(after) {
		next = s.advance(next)
	}

	return next, nil
}

func (s Schedule) advance(current time.Time) time.Time {
	switch s.Unit {
	case RecurrenceUnitDay:
		return current.AddDate(0, 0, s.Every)
	case RecurrenceUnitWeek:
		return current.AddDate(0, 0, s.Every*7)
	case RecurrenceUnitMonth:
		return current.AddDate(0, s.Every, 0)
	case RecurrenceUnitQuarter:
		return current.AddDate(0, s.Every*3, 0)
	case RecurrenceUnitYear:
		return current.AddDate(s.Every, 0, 0)
	default:
		return current
	}
}

func (s Schedule) validate() error {
	if s.StartDate.IsZero() {
		return ErrStartDateRequired
	}
	if s.Every <= 0 {
		return ErrIntervalRequired
	}
	switch s.Unit {
	case RecurrenceUnitDay, RecurrenceUnitWeek, RecurrenceUnitMonth, RecurrenceUnitQuarter, RecurrenceUnitYear:
		return nil
	default:
		return ErrInvalidRecurrenceUnit
	}
}
