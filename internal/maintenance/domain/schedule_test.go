package domain_test

import (
	"time"

	maintenanceDomain "zensor-server/internal/maintenance/domain"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Schedule", func() {
	var startDate time.Time

	BeforeEach(func() {
		startDate = time.Date(2026, 9, 1, 9, 0, 0, 0, time.UTC)
	})

	Context("Next", func() {
		When("unit is day", func() {
			It("returns the start date when the reference is before it", func() {
				schedule := maintenanceDomain.Schedule{
					StartDate: startDate,
					Every:     1,
					Unit:      maintenanceDomain.RecurrenceUnitDay,
				}

				next, err := schedule.Next(startDate.Add(-time.Hour))
				Expect(err).NotTo(HaveOccurred())
				Expect(next).To(Equal(startDate))
			})

			It("returns the next daily occurrence after the reference", func() {
				schedule := maintenanceDomain.Schedule{
					StartDate: startDate,
					Every:     1,
					Unit:      maintenanceDomain.RecurrenceUnitDay,
				}

				next, err := schedule.Next(startDate.AddDate(0, 0, 2))
				Expect(err).NotTo(HaveOccurred())
				Expect(next).To(Equal(startDate.AddDate(0, 0, 3)))
			})
		})

		When("unit is week", func() {
			It("advances by seven days per interval", func() {
				schedule := maintenanceDomain.Schedule{
					StartDate: startDate,
					Every:     2,
					Unit:      maintenanceDomain.RecurrenceUnitWeek,
				}

				next, err := schedule.Next(startDate.AddDate(0, 0, 13))
				Expect(err).NotTo(HaveOccurred())
				Expect(next).To(Equal(startDate.AddDate(0, 0, 14)))
			})
		})

		When("unit is month", func() {
			It("advances by months per interval", func() {
				schedule := maintenanceDomain.Schedule{
					StartDate: startDate,
					Every:     2,
					Unit:      maintenanceDomain.RecurrenceUnitMonth,
				}

				next, err := schedule.Next(time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC))
				Expect(err).NotTo(HaveOccurred())
				Expect(next).To(Equal(time.Date(2027, 1, 1, 9, 0, 0, 0, time.UTC)))
			})
		})

		When("unit is quarter", func() {
			It("advances by three months per interval", func() {
				schedule := maintenanceDomain.Schedule{
					StartDate: startDate,
					Every:     1,
					Unit:      maintenanceDomain.RecurrenceUnitQuarter,
				}

				next, err := schedule.Next(time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC))
				Expect(err).NotTo(HaveOccurred())
				Expect(next).To(Equal(time.Date(2026, 12, 1, 9, 0, 0, 0, time.UTC)))
			})
		})

		When("unit is year", func() {
			It("advances by years per interval", func() {
				schedule := maintenanceDomain.Schedule{
					StartDate: startDate,
					Every:     4,
					Unit:      maintenanceDomain.RecurrenceUnitYear,
				}

				next, err := schedule.Next(time.Date(2029, 6, 1, 0, 0, 0, 0, time.UTC))
				Expect(err).NotTo(HaveOccurred())
				Expect(next).To(Equal(time.Date(2030, 9, 1, 9, 0, 0, 0, time.UTC)))
			})
		})

		It("preserves the time of day of the start date", func() {
			schedule := maintenanceDomain.Schedule{
				StartDate: time.Date(2026, 9, 1, 14, 30, 0, 0, time.UTC),
				Every:     1,
				Unit:      maintenanceDomain.RecurrenceUnitWeek,
			}

			next, err := schedule.Next(time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC))
			Expect(err).NotTo(HaveOccurred())
			Expect(next.Hour()).To(Equal(14))
			Expect(next.Minute()).To(Equal(30))
		})
	})

	Context("validation", func() {
		It("returns an error when the start date is zero", func() {
			schedule := maintenanceDomain.Schedule{
				Every: 1,
				Unit:  maintenanceDomain.RecurrenceUnitMonth,
			}

			_, err := schedule.Next(time.Now())
			Expect(err).To(Equal(maintenanceDomain.ErrStartDateRequired))
		})

		It("returns an error when the interval is not positive", func() {
			schedule := maintenanceDomain.Schedule{
				StartDate: startDate,
				Every:     0,
				Unit:      maintenanceDomain.RecurrenceUnitMonth,
			}

			_, err := schedule.Next(time.Now())
			Expect(err).To(Equal(maintenanceDomain.ErrIntervalRequired))
		})

		It("returns an error when the unit is invalid", func() {
			schedule := maintenanceDomain.Schedule{
				StartDate: startDate,
				Every:     1,
				Unit:      maintenanceDomain.RecurrenceUnit("fortnight"),
			}

			_, err := schedule.Next(time.Now())
			Expect(err).To(Equal(maintenanceDomain.ErrInvalidRecurrenceUnit))
		})
	})
})
