package domain_test

import (
	"time"

	medicinesDomain "zensor-server/internal/medicines/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("MedicineSchedule", func() {
	var startAt time.Time

	ginkgo.BeforeEach(func() {
		startAt = time.Date(2026, time.March, 1, 8, 0, 0, 0, time.UTC)
	})

	ginkgo.Context("Validate", func() {
		ginkgo.When("the start date is missing", func() {
			ginkgo.It("should fail", func() {
				schedule := medicinesDomain.MedicineSchedule{Every: 8, Unit: medicinesDomain.IntervalUnitHour}
				gomega.Expect(schedule.Validate()).To(gomega.MatchError(medicinesDomain.ErrStartAtRequired))
			})
		})

		ginkgo.When("the interval is not positive", func() {
			ginkgo.It("should fail", func() {
				schedule := medicinesDomain.MedicineSchedule{StartAt: startAt, Every: 0, Unit: medicinesDomain.IntervalUnitHour}
				gomega.Expect(schedule.Validate()).To(gomega.MatchError(medicinesDomain.ErrIntervalRequired))
			})
		})

		ginkgo.When("the unit is unknown", func() {
			ginkgo.It("should fail", func() {
				schedule := medicinesDomain.MedicineSchedule{StartAt: startAt, Every: 1, Unit: "fortnight"}
				gomega.Expect(schedule.Validate()).To(gomega.MatchError(medicinesDomain.ErrInvalidIntervalUnit))
			})
		})

		ginkgo.When("the schedule is well formed", func() {
			ginkgo.It("should pass for both units", func() {
				hourly := medicinesDomain.MedicineSchedule{StartAt: startAt, Every: 8, Unit: medicinesDomain.IntervalUnitHour}
				daily := medicinesDomain.MedicineSchedule{StartAt: startAt, Every: 1, Unit: medicinesDomain.IntervalUnitDay}
				gomega.Expect(hourly.Validate()).To(gomega.Succeed())
				gomega.Expect(daily.Validate()).To(gomega.Succeed())
			})
		})
	})

	ginkgo.Context("OccurrenceAt", func() {
		ginkgo.When("the unit is hour", func() {
			ginkgo.It("should return the start date for index zero", func() {
				schedule := medicinesDomain.MedicineSchedule{StartAt: startAt, Every: 8, Unit: medicinesDomain.IntervalUnitHour}
				gomega.Expect(schedule.OccurrenceAt(0)).To(gomega.Equal(startAt))
			})

			ginkgo.It("should multiply the interval by the index", func() {
				schedule := medicinesDomain.MedicineSchedule{StartAt: startAt, Every: 8, Unit: medicinesDomain.IntervalUnitHour}
				gomega.Expect(schedule.OccurrenceAt(3)).To(gomega.Equal(startAt.Add(24 * time.Hour)))
			})
		})

		ginkgo.When("the unit is day", func() {
			ginkgo.It("should add calendar days", func() {
				schedule := medicinesDomain.MedicineSchedule{StartAt: startAt, Every: 2, Unit: medicinesDomain.IntervalUnitDay}
				gomega.Expect(schedule.OccurrenceAt(3)).To(gomega.Equal(startAt.AddDate(0, 0, 6)))
			})
		})
	})

	ginkgo.Context("Next", func() {
		var schedule medicinesDomain.MedicineSchedule

		ginkgo.BeforeEach(func() {
			schedule = medicinesDomain.MedicineSchedule{StartAt: startAt, Every: 8, Unit: medicinesDomain.IntervalUnitHour}
		})

		ginkgo.When("the reference instant is before the start date", func() {
			ginkgo.It("should return the start date", func() {
				next, err := schedule.Next(startAt.Add(-1 * time.Hour))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(next).To(gomega.Equal(startAt))
			})
		})

		ginkgo.When("the reference instant lands exactly on an occurrence", func() {
			ginkgo.It("should return the following one, never the same", func() {
				next, err := schedule.Next(startAt)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(next).To(gomega.Equal(startAt.Add(8 * time.Hour)))

				next, err = schedule.Next(startAt.Add(16 * time.Hour))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(next).To(gomega.Equal(startAt.Add(24 * time.Hour)))
			})
		})

		ginkgo.When("the reference instant falls between two occurrences", func() {
			ginkgo.It("should round up to the next one", func() {
				next, err := schedule.Next(startAt.Add(20 * time.Hour))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(next).To(gomega.Equal(startAt.Add(24 * time.Hour)))
			})
		})

		ginkgo.When("the schedule has been running for months", func() {
			ginkgo.It("should resolve arithmetically rather than by stepping", func() {
				every4h := medicinesDomain.MedicineSchedule{StartAt: startAt, Every: 4, Unit: medicinesDomain.IntervalUnitHour}
				after := startAt.Add(90 * 24 * time.Hour)

				next, err := every4h.Next(after)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(next).To(gomega.Equal(after.Add(4 * time.Hour)))
			})
		})

		ginkgo.When("the schedule is invalid", func() {
			ginkgo.It("should propagate the validation error", func() {
				invalid := medicinesDomain.MedicineSchedule{StartAt: startAt, Every: -1, Unit: medicinesDomain.IntervalUnitHour}
				_, err := invalid.Next(startAt)
				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrIntervalRequired))
			})
		})
	})

	ginkgo.Context("OccurrenceIndicesBetween", func() {
		var schedule medicinesDomain.MedicineSchedule

		ginkgo.BeforeEach(func() {
			schedule = medicinesDomain.MedicineSchedule{StartAt: startAt, Every: 8, Unit: medicinesDomain.IntervalUnitHour}
		})

		ginkgo.When("the window covers the first occurrences", func() {
			ginkgo.It("should return them in order", func() {
				indices, err := schedule.OccurrenceIndicesBetween(startAt, startAt.Add(20*time.Hour))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(indices).To(gomega.Equal([]int{0, 1, 2}))
			})
		})

		ginkgo.When("an occurrence lands exactly on a window boundary", func() {
			ginkgo.It("should include both ends, the interval is closed", func() {
				indices, err := schedule.OccurrenceIndicesBetween(startAt.Add(8*time.Hour), startAt.Add(24*time.Hour))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(indices).To(gomega.Equal([]int{1, 2, 3}))
			})
		})

		ginkgo.When("the window lies entirely before the start date", func() {
			ginkgo.It("should return nothing", func() {
				indices, err := schedule.OccurrenceIndicesBetween(startAt.Add(-48*time.Hour), startAt.Add(-1*time.Hour))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(indices).To(gomega.BeEmpty())
			})
		})

		ginkgo.When("the window falls between two occurrences", func() {
			ginkgo.It("should return nothing", func() {
				indices, err := schedule.OccurrenceIndicesBetween(startAt.Add(1*time.Hour), startAt.Add(7*time.Hour))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(indices).To(gomega.BeEmpty())
			})
		})

		ginkgo.When("the window would produce more occurrences than the cap", func() {
			ginkgo.It("should fail rather than return an unbounded batch", func() {
				hourly := medicinesDomain.MedicineSchedule{StartAt: startAt, Every: 1, Unit: medicinesDomain.IntervalUnitHour}
				_, err := hourly.OccurrenceIndicesBetween(startAt, startAt.Add(500*time.Hour))
				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrScheduleWindowTooLarge))
			})
		})

		ginkgo.When("the window is inverted", func() {
			ginkgo.It("should return nothing", func() {
				indices, err := schedule.OccurrenceIndicesBetween(startAt.Add(24*time.Hour), startAt)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(indices).To(gomega.BeEmpty())
			})
		})
	})

	ginkgo.Context("daylight saving transitions", func() {
		var santiago *time.Location

		ginkgo.BeforeEach(func() {
			var err error
			santiago, err = time.LoadLocation("America/Santiago")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.When("the unit is day", func() {
			ginkgo.It("should preserve the wall clock time across the transition", func() {
				// Chile moves the clocks forward on 2026-09-06.
				beforeTransition := time.Date(2026, time.September, 5, 8, 0, 0, 0, santiago)
				schedule := medicinesDomain.MedicineSchedule{
					StartAt: beforeTransition,
					Every:   1,
					Unit:    medicinesDomain.IntervalUnitDay,
				}

				afterTransition := schedule.OccurrenceAt(2)

				gomega.Expect(afterTransition.In(santiago).Hour()).To(gomega.Equal(8))
				gomega.Expect(afterTransition.In(santiago).Day()).To(gomega.Equal(7))
			})
		})

		ginkgo.When("the unit is hour", func() {
			ginkgo.It("should preserve the real elapsed interval, letting the wall clock shift", func() {
				beforeTransition := time.Date(2026, time.September, 5, 8, 0, 0, 0, santiago)
				schedule := medicinesDomain.MedicineSchedule{
					StartAt: beforeTransition,
					Every:   24,
					Unit:    medicinesDomain.IntervalUnitHour,
				}

				afterTransition := schedule.OccurrenceAt(2)

				gomega.Expect(afterTransition.Sub(beforeTransition)).To(gomega.Equal(48 * time.Hour))
				gomega.Expect(afterTransition.In(santiago).Hour()).To(gomega.Equal(9))
			})
		})
	})
})
