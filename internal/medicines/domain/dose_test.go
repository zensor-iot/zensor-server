package domain_test

import (
	"time"

	medicinesDomain "zensor-server/internal/medicines/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Dose", func() {
	var scheduledAt time.Time

	newDose := func() medicinesDomain.Dose {
		dose, err := medicinesDomain.NewDoseBuilder().
			WithTreatmentID("treatment-1").
			WithSequenceNumber(0).
			WithScheduledAt(scheduledAt).
			WithDose(15, medicinesDomain.DoseUnitDrop).
			Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return dose
	}

	ginkgo.BeforeEach(func() {
		scheduledAt = time.Date(2026, time.March, 1, 8, 0, 0, 0, time.UTC)
	})

	ginkgo.Context("Build", func() {
		ginkgo.When("all required values are present", func() {
			ginkgo.It("should start pending", func() {
				dose := newDose()
				gomega.Expect(dose.Status).To(gomega.Equal(medicinesDomain.DoseStatusPending))
				gomega.Expect(dose.IsResolved()).To(gomega.BeFalse())
				gomega.Expect(dose.HasReminderBeenSent()).To(gomega.BeFalse())
			})
		})

		ginkgo.When("the treatment is missing", func() {
			ginkgo.It("should fail", func() {
				_, err := medicinesDomain.NewDoseBuilder().
					WithScheduledAt(scheduledAt).
					WithDose(1, medicinesDomain.DoseUnitPill).
					Build()
				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrTreatmentIDRequired))
			})
		})

		ginkgo.When("the scheduled date is missing", func() {
			ginkgo.It("should fail", func() {
				_, err := medicinesDomain.NewDoseBuilder().
					WithTreatmentID("treatment-1").
					WithDose(1, medicinesDomain.DoseUnitPill).
					Build()
				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrScheduledAtRequired))
			})
		})

		ginkgo.When("the sequence number is negative", func() {
			ginkgo.It("should fail", func() {
				_, err := medicinesDomain.NewDoseBuilder().
					WithTreatmentID("treatment-1").
					WithSequenceNumber(-1).
					WithScheduledAt(scheduledAt).
					WithDose(1, medicinesDomain.DoseUnitPill).
					Build()
				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrSequenceNumberInvalid))
			})
		})
	})

	ginkgo.Context("MarkAdministered", func() {
		ginkgo.When("the dose is pending", func() {
			ginkgo.It("should record who resolved it and when", func() {
				dose := newDose()
				notes := medicinesDomain.DoseNotes("con comida")

				err := dose.MarkAdministered("user-1", &notes)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(dose.Status).To(gomega.Equal(medicinesDomain.DoseStatusAdministered))
				gomega.Expect(dose.IsResolved()).To(gomega.BeTrue())
				gomega.Expect(*dose.ResolvedBy).To(gomega.BeEquivalentTo("user-1"))
				gomega.Expect(dose.ResolvedAt).NotTo(gomega.BeNil())
				gomega.Expect(*dose.Notes).To(gomega.Equal(notes))
			})
		})

		ginkgo.When("the dose was already resolved", func() {
			ginkgo.It("should refuse to resolve it twice", func() {
				dose := newDose()
				gomega.Expect(dose.MarkAdministered("user-1", nil)).To(gomega.Succeed())

				err := dose.MarkAdministered("user-2", nil)

				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrDoseAlreadyResolved))
				gomega.Expect(*dose.ResolvedBy).To(gomega.BeEquivalentTo("user-1"))
			})
		})

		ginkgo.When("the dose is scheduled in the future", func() {
			ginkgo.It("should still allow it, giving a dose early is legitimate", func() {
				dose, err := medicinesDomain.NewDoseBuilder().
					WithTreatmentID("treatment-1").
					WithScheduledAt(time.Now().Add(1*time.Hour)).
					WithDose(1, medicinesDomain.DoseUnitPill).
					Build()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				gomega.Expect(dose.MarkAdministered("user-1", nil)).To(gomega.Succeed())
			})
		})
	})

	ginkgo.Context("MarkSkipped", func() {
		ginkgo.When("the dose is pending", func() {
			ginkgo.It("should mark it skipped", func() {
				dose := newDose()
				gomega.Expect(dose.MarkSkipped("user-1", nil)).To(gomega.Succeed())
				gomega.Expect(dose.Status).To(gomega.Equal(medicinesDomain.DoseStatusSkipped))
			})
		})

		ginkgo.When("the dose was already administered", func() {
			ginkgo.It("should refuse", func() {
				dose := newDose()
				gomega.Expect(dose.MarkAdministered("user-1", nil)).To(gomega.Succeed())
				gomega.Expect(dose.MarkSkipped("user-1", nil)).To(gomega.MatchError(medicinesDomain.ErrDoseAlreadyResolved))
			})
		})
	})

	ginkgo.Context("IsOverdue", func() {
		ginkgo.When("the dose is pending and past its time", func() {
			ginkgo.It("should be overdue", func() {
				dose := newDose()
				gomega.Expect(dose.IsOverdue(scheduledAt.Add(1 * time.Minute))).To(gomega.BeTrue())
			})
		})

		ginkgo.When("the dose is pending and not yet due", func() {
			ginkgo.It("should not be overdue", func() {
				dose := newDose()
				gomega.Expect(dose.IsOverdue(scheduledAt.Add(-1 * time.Minute))).To(gomega.BeFalse())
			})
		})

		ginkgo.When("the dose was resolved", func() {
			ginkgo.It("should not be overdue", func() {
				dose := newDose()
				gomega.Expect(dose.MarkSkipped("user-1", nil)).To(gomega.Succeed())
				gomega.Expect(dose.IsOverdue(scheduledAt.Add(72 * time.Hour))).To(gomega.BeFalse())
			})
		})
	})

	ginkgo.Context("MarkReminderSent", func() {
		ginkgo.When("a reminder is emitted", func() {
			ginkgo.It("should stamp the dose so it is not sent twice", func() {
				dose := newDose()
				dose.MarkReminderSent()
				gomega.Expect(dose.HasReminderBeenSent()).To(gomega.BeTrue())
			})
		})
	})
})
