package domain

import "time"

// IntervalUnit is the unit of a treatment's recurrence interval.
type IntervalUnit string

const (
	IntervalUnitHour IntervalUnit = "hour"
	IntervalUnitDay  IntervalUnit = "day"
)

// _maxOccurrencesPerWindow bounds how many doses a single query window may
// produce, so an "every hour, indefinitely" treatment cannot make the worker
// materialise an unbounded batch.
const _maxOccurrencesPerWindow = 200

// MedicineSchedule describes when the doses of a treatment fall due: the first
// dose is at StartAt and every following one is Every units later.
//
// Occurrences are computed arithmetically rather than by stepping forward one
// interval at a time, so a lookup stays cheap for a treatment that has been
// running for months at an hourly interval.
//
// The two units have deliberately different arithmetic:
//
//   - IntervalUnitHour adds a fixed duration. "Every 8 hours" stays 8 real
//     hours across a daylight saving transition, because a dosing interval is
//     a pharmacological quantity and does not observe the clock.
//   - IntervalUnitDay adds calendar days, preserving the wall clock time. The
//     08:00 pill stays at 08:00 after the clocks change, which is what the
//     person giving it expects.
type MedicineSchedule struct {
	StartAt time.Time    `json:"start_at"`
	Every   int          `json:"every"`
	Unit    IntervalUnit `json:"unit"`
}

// Validate reports whether the schedule can produce occurrences.
func (s MedicineSchedule) Validate() error {
	if s.StartAt.IsZero() {
		return ErrStartAtRequired
	}
	if s.Every <= 0 {
		return ErrIntervalRequired
	}
	switch s.Unit {
	case IntervalUnitHour, IntervalUnitDay:
		return nil
	default:
		return ErrInvalidIntervalUnit
	}
}

// OccurrenceAt returns the date of the nth dose, counting from zero, so
// OccurrenceAt(0) is StartAt.
func (s MedicineSchedule) OccurrenceAt(n int) time.Time {
	switch s.Unit {
	case IntervalUnitDay:
		return s.StartAt.AddDate(0, 0, n*s.Every)
	case IntervalUnitHour:
		return s.StartAt.Add(time.Duration(n) * s.interval())
	default:
		return s.StartAt
	}
}

// Next returns the first occurrence strictly after the given instant.
func (s MedicineSchedule) Next(after time.Time) (time.Time, error) {
	if err := s.Validate(); err != nil {
		return time.Time{}, err
	}

	if after.Before(s.StartAt) {
		return s.StartAt, nil
	}

	n := int(after.Sub(s.StartAt) / s.interval())
	next := s.OccurrenceAt(n)
	// Calendar arithmetic and the duration estimate can disagree by an hour
	// around a daylight saving transition, so settle the boundary exactly.
	for !next.After(after) {
		n++
		next = s.OccurrenceAt(n)
	}

	return next, nil
}

// OccurrenceIndicesBetween returns the indices of every occurrence falling in
// the closed interval [from, to].
func (s MedicineSchedule) OccurrenceIndicesBetween(from, to time.Time) ([]int, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	if to.Before(s.StartAt) || to.Before(from) {
		return nil, nil
	}

	low := 0
	if from.After(s.StartAt) {
		low = int(from.Sub(s.StartAt) / s.interval())
		for s.OccurrenceAt(low).Before(from) {
			low++
		}
		for low > 0 && !s.OccurrenceAt(low-1).Before(from) {
			low--
		}
	}

	high := int(to.Sub(s.StartAt) / s.interval())
	for s.OccurrenceAt(high).After(to) {
		high--
	}
	for !s.OccurrenceAt(high + 1).After(to) {
		high++
	}

	if high < low {
		return nil, nil
	}
	if high-low+1 > _maxOccurrencesPerWindow {
		return nil, ErrScheduleWindowTooLarge
	}

	indices := make([]int, 0, high-low+1)
	for n := low; n <= high; n++ {
		indices = append(indices, n)
	}

	return indices, nil
}

// interval is the nominal length of one step. For IntervalUnitDay it is only
// an estimate used to seed the arithmetic; OccurrenceAt does the calendar
// correct work.
func (s MedicineSchedule) interval() time.Duration {
	switch s.Unit {
	case IntervalUnitDay:
		return time.Duration(s.Every) * 24 * time.Hour
	case IntervalUnitHour:
		return time.Duration(s.Every) * time.Hour
	default:
		return time.Duration(s.Every) * time.Hour
	}
}
