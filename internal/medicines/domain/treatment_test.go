package domain_test

import (
	"time"

	medicinesDomain "zensor-server/internal/medicines/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Treatment", func() {
	var (
		startAt  time.Time
		schedule medicinesDomain.MedicineSchedule
	)

	ginkgo.BeforeEach(func() {
		startAt = time.Date(2026, time.March, 1, 8, 0, 0, 0, time.UTC)
		schedule = medicinesDomain.MedicineSchedule{
			StartAt: startAt,
			Every:   8,
			Unit:    medicinesDomain.IntervalUnitHour,
		}
	})

	newTreatment := func(mutate func(*medicinesDomain.Treatment)) medicinesDomain.Treatment {
		treatment, err := medicinesDomain.NewTreatmentBuilder().
			WithTenantID("tenant-1").
			WithPatientID("patient-1").
			WithMedicineName("Amoxicilina").
			WithDose(15, medicinesDomain.DoseUnitDrop).
			WithSchedule(schedule).
			Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		if mutate != nil {
			mutate(&treatment)
		}
		return treatment
	}

	ginkgo.Context("Build", func() {
		ginkgo.When("all required values are present", func() {
			ginkgo.It("should start active with version one", func() {
				treatment := newTreatment(nil)
				gomega.Expect(treatment.IsActive).To(gomega.BeTrue())
				gomega.Expect(treatment.Version).To(gomega.BeEquivalentTo(1))
				gomega.Expect(treatment.ID).NotTo(gomega.BeEmpty())
			})
		})

		ginkgo.When("the patient is missing", func() {
			ginkgo.It("should fail", func() {
				_, err := medicinesDomain.NewTreatmentBuilder().
					WithTenantID("tenant-1").
					WithMedicineName("Amoxicilina").
					WithDose(1, medicinesDomain.DoseUnitPill).
					WithSchedule(schedule).
					Build()
				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrPatientIDRequired))
			})
		})

		ginkgo.When("the quantity is not positive", func() {
			ginkgo.It("should fail", func() {
				_, err := medicinesDomain.NewTreatmentBuilder().
					WithTenantID("tenant-1").
					WithPatientID("patient-1").
					WithMedicineName("Amoxicilina").
					WithDose(0, medicinesDomain.DoseUnitPill).
					WithSchedule(schedule).
					Build()
				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrDoseQuantityRequired))
			})
		})

		ginkgo.When("the unit is unknown", func() {
			ginkgo.It("should fail", func() {
				_, err := medicinesDomain.NewTreatmentBuilder().
					WithTenantID("tenant-1").
					WithPatientID("patient-1").
					WithMedicineName("Amoxicilina").
					WithDose(1, "sachet").
					WithSchedule(schedule).
					Build()
				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrInvalidDoseUnit))
			})
		})

		ginkgo.When("the total doses limit is not positive", func() {
			ginkgo.It("should fail", func() {
				zero := 0
				_, err := medicinesDomain.NewTreatmentBuilder().
					WithTenantID("tenant-1").
					WithPatientID("patient-1").
					WithMedicineName("Amoxicilina").
					WithDose(1, medicinesDomain.DoseUnitPill).
					WithSchedule(schedule).
					WithTotalDoses(&zero).
					Build()
				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrTotalDosesInvalid))
			})
		})

		ginkgo.When("the end date is not after the schedule start", func() {
			ginkgo.It("should fail", func() {
				endAt := startAt.Add(-1 * time.Hour)
				_, err := medicinesDomain.NewTreatmentBuilder().
					WithTenantID("tenant-1").
					WithPatientID("patient-1").
					WithMedicineName("Amoxicilina").
					WithDose(1, medicinesDomain.DoseUnitPill).
					WithSchedule(schedule).
					WithEndAt(&endAt).
					Build()
				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrEndAtBeforeStartAt))
			})
		})
	})

	ginkgo.Context("DoseText", func() {
		ginkgo.When("the quantity is greater than one", func() {
			ginkgo.It("should pluralise the unit", func() {
				treatment := newTreatment(nil)
				gomega.Expect(treatment.DoseText()).To(gomega.Equal("15 gotas"))
			})
		})

		ginkgo.When("the quantity is exactly one", func() {
			ginkgo.It("should use the singular", func() {
				treatment := newTreatment(func(t *medicinesDomain.Treatment) {
					t.Quantity = 1
					t.Unit = medicinesDomain.DoseUnitPill
				})
				gomega.Expect(treatment.DoseText()).To(gomega.Equal("1 pastilla"))
			})
		})

		ginkgo.When("the quantity is fractional", func() {
			ginkgo.It("should render without trailing zeros", func() {
				treatment := newTreatment(func(t *medicinesDomain.Treatment) {
					t.Quantity = 0.5
					t.Unit = medicinesDomain.DoseUnitSpoonful
				})
				gomega.Expect(treatment.DoseText()).To(gomega.Equal("0.5 cucharada"))
			})
		})
	})

	ginkgo.Context("EffectiveEndAt", func() {
		ginkgo.When("the treatment is indefinite", func() {
			ginkgo.It("should return nothing", func() {
				treatment := newTreatment(nil)
				gomega.Expect(treatment.EffectiveEndAt()).To(gomega.BeNil())
			})
		})

		ginkgo.When("only a dose count is set", func() {
			ginkgo.It("should return the date of the last dose", func() {
				total := 3
				treatment := newTreatment(func(t *medicinesDomain.Treatment) { t.TotalDoses = &total })
				gomega.Expect(*treatment.EffectiveEndAt()).To(gomega.Equal(startAt.Add(16 * time.Hour)))
			})
		})

		ginkgo.When("both limits are set", func() {
			ginkgo.It("should return whichever comes first", func() {
				total := 10
				endAt := startAt.Add(20 * time.Hour)
				treatment := newTreatment(func(t *medicinesDomain.Treatment) {
					t.TotalDoses = &total
					t.EndAt = &endAt
				})
				gomega.Expect(*treatment.EffectiveEndAt()).To(gomega.Equal(endAt))

				earlierCount := 2
				treatment.TotalDoses = &earlierCount
				gomega.Expect(*treatment.EffectiveEndAt()).To(gomega.Equal(startAt.Add(8 * time.Hour)))
			})
		})
	})

	ginkgo.Context("DueOccurrences", func() {
		ginkgo.When("the treatment is indefinite", func() {
			ginkgo.It("should return every occurrence in the window", func() {
				treatment := newTreatment(nil)
				occurrences, err := treatment.DueOccurrences(startAt, startAt.Add(24*time.Hour))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(occurrences).To(gomega.HaveLen(4))
				gomega.Expect(occurrences[0].SequenceNumber).To(gomega.Equal(0))
				gomega.Expect(occurrences[3].ScheduledAt).To(gomega.Equal(startAt.Add(24 * time.Hour)))
			})
		})

		ginkgo.When("a total dose count is set", func() {
			ginkgo.It("should stop at the last dose even if the window is wider", func() {
				total := 3
				treatment := newTreatment(func(t *medicinesDomain.Treatment) { t.TotalDoses = &total })

				occurrences, err := treatment.DueOccurrences(startAt, startAt.Add(72*time.Hour))

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(occurrences).To(gomega.HaveLen(3))
				gomega.Expect(occurrences[2].SequenceNumber).To(gomega.Equal(2))
			})
		})

		ginkgo.When("an end date is set", func() {
			ginkgo.It("should stop at the end date", func() {
				endAt := startAt.Add(17 * time.Hour)
				treatment := newTreatment(func(t *medicinesDomain.Treatment) { t.EndAt = &endAt })

				occurrences, err := treatment.DueOccurrences(startAt, startAt.Add(72*time.Hour))

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(occurrences).To(gomega.HaveLen(3))
			})
		})

		ginkgo.When("both limits are set", func() {
			ginkgo.It("should apply whichever cuts first", func() {
				total := 2
				endAt := startAt.Add(72 * time.Hour)
				treatment := newTreatment(func(t *medicinesDomain.Treatment) {
					t.TotalDoses = &total
					t.EndAt = &endAt
				})

				occurrences, err := treatment.DueOccurrences(startAt, startAt.Add(72*time.Hour))

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(occurrences).To(gomega.HaveLen(2))
			})
		})

		ginkgo.When("the treatment is inactive", func() {
			ginkgo.It("should return nothing", func() {
				treatment := newTreatment(func(t *medicinesDomain.Treatment) { t.Deactivate() })
				occurrences, err := treatment.DueOccurrences(startAt, startAt.Add(72*time.Hour))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(occurrences).To(gomega.BeEmpty())
			})
		})

		ginkgo.When("the treatment is deleted", func() {
			ginkgo.It("should return nothing", func() {
				treatment := newTreatment(func(t *medicinesDomain.Treatment) { t.SoftDelete() })
				occurrences, err := treatment.DueOccurrences(startAt, startAt.Add(72*time.Hour))
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(occurrences).To(gomega.BeEmpty())
			})
		})
	})
})
