package persistence_test

import (
	"context"
	"time"
	"zensor-server/internal/infra/sql"

	medicinesDomain "zensor-server/internal/medicines/domain"
	medicinesPersistence "zensor-server/internal/medicines/persistence"
	medicinesPersistenceInternal "zensor-server/internal/medicines/persistence/internal"
	medicinesUsecases "zensor-server/internal/medicines/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("DoseRepository", func() {
	var (
		orm                 sql.ORM
		doseRepo            medicinesUsecases.DoseRepository
		treatmentRepo       medicinesUsecases.TreatmentRepository
		patientRepo         medicinesUsecases.PatientRepository
		ctx                 context.Context
		patient             medicinesDomain.Patient
		treatment           medicinesDomain.Treatment
		scheduleStart       time.Time
		buildAndStoreEntity func(seq int, at time.Time) medicinesDomain.Dose
	)

	ginkgo.BeforeEach(func() {
		var err error
		orm, err = sql.NewMemoryORM("migrations")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		patientRepo, err = medicinesPersistence.NewPatientRepository(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		treatmentRepo, err = medicinesPersistence.NewTreatmentRepository(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		doseRepo, err = medicinesPersistence.NewDoseRepository(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ctx = context.Background()
		scheduleStart = time.Now().UTC().Round(0).Add(1 * time.Hour)

		// The in-memory ORM is a process wide singleton, so rows survive
		// between specs. These window queries span every treatment, so the
		// tables have to start empty. Cleanup runs after the repositories are
		// built, because that is what creates the tables.
		for _, model := range []any{
			&medicinesPersistenceInternal.Dose{},
			&medicinesPersistenceInternal.Treatment{},
			&medicinesPersistenceInternal.Patient{},
		} {
			gomega.Expect(orm.WithContext(ctx).Unscoped().Where("1 = 1").Delete(model).Error()).To(gomega.Succeed())
		}

		patient, err = medicinesDomain.NewPatientBuilder().
			WithTenantID("tenant-1").
			WithName("Luna").
			WithKind(medicinesDomain.PatientKindAnimal).
			Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(patientRepo.Create(ctx, patient)).To(gomega.Succeed())

		treatment, err = medicinesDomain.NewTreatmentBuilder().
			WithTenantID("tenant-1").
			WithPatientID(patient.ID.String()).
			WithMedicineName("Amoxicilina").
			WithDose(15, medicinesDomain.DoseUnitDrop).
			WithSchedule(medicinesDomain.MedicineSchedule{
				StartAt: scheduleStart,
				Every:   8,
				Unit:    medicinesDomain.IntervalUnitHour,
			}).
			Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(treatmentRepo.Create(ctx, treatment)).To(gomega.Succeed())

		buildAndStoreEntity = func(seq int, at time.Time) medicinesDomain.Dose {
			dose, buildErr := medicinesDomain.NewDoseBuilder().
				WithTreatmentID(treatment.ID.String()).
				WithSequenceNumber(seq).
				WithScheduledAt(at).
				WithDose(15, medicinesDomain.DoseUnitDrop).
				Build()
			gomega.Expect(buildErr).NotTo(gomega.HaveOccurred())
			gomega.Expect(doseRepo.CreateBatch(ctx, []medicinesDomain.Dose{dose})).To(gomega.Succeed())
			return dose
		}
	})

	ginkgo.Context("CreateBatch", func() {
		ginkgo.When("the same treatment and sequence number is inserted twice", func() {
			ginkgo.It("should be rejected by the unique index", func() {
				buildAndStoreEntity(0, scheduleStart)

				duplicate, err := medicinesDomain.NewDoseBuilder().
					WithTreatmentID(treatment.ID.String()).
					WithSequenceNumber(0).
					WithScheduledAt(scheduleStart).
					WithDose(15, medicinesDomain.DoseUnitDrop).
					Build()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				err = doseRepo.CreateBatch(ctx, []medicinesDomain.Dose{duplicate})

				gomega.Expect(err).To(gomega.HaveOccurred())
			})
		})

		ginkgo.When("the batch is empty", func() {
			ginkgo.It("should be a no-op", func() {
				gomega.Expect(doseRepo.CreateBatch(ctx, nil)).To(gomega.Succeed())
			})
		})
	})

	ginkgo.Context("FindByTreatmentInWindow", func() {
		ginkgo.When("doses fall inside and outside the window", func() {
			ginkgo.It("should return only those inside", func() {
				buildAndStoreEntity(0, scheduleStart)
				buildAndStoreEntity(1, scheduleStart.Add(8*time.Hour))
				buildAndStoreEntity(9, scheduleStart.Add(72*time.Hour))

				doses, err := doseRepo.FindByTreatmentInWindow(ctx, treatment.ID, scheduleStart, scheduleStart.Add(24*time.Hour))

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(doses).To(gomega.HaveLen(2))
			})
		})
	})

	ginkgo.Context("FindDueForReminder", func() {
		var from, to time.Time

		ginkgo.BeforeEach(func() {
			from = scheduleStart.Add(-1 * time.Minute)
			to = scheduleStart.Add(1 * time.Minute)
		})

		ginkgo.When("a pending dose falls in the window", func() {
			ginkgo.It("should return it with its treatment and patient", func() {
				buildAndStoreEntity(0, scheduleStart)

				entries, err := doseRepo.FindDueForReminder(ctx, from, to, 10)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(entries).To(gomega.HaveLen(1))
				gomega.Expect(entries[0].Treatment.MedicineName).To(gomega.BeEquivalentTo("Amoxicilina"))
				gomega.Expect(entries[0].Patient.Name).To(gomega.BeEquivalentTo("Luna"))
			})
		})

		ginkgo.When("the dose was already resolved", func() {
			ginkgo.It("should be excluded", func() {
				dose := buildAndStoreEntity(0, scheduleStart)
				gomega.Expect(dose.MarkAdministered("user-1", nil)).To(gomega.Succeed())
				gomega.Expect(doseRepo.Update(ctx, dose)).To(gomega.Succeed())

				entries, err := doseRepo.FindDueForReminder(ctx, from, to, 10)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(entries).To(gomega.BeEmpty())
			})
		})

		ginkgo.When("a reminder was already sent", func() {
			ginkgo.It("should be excluded, so it is never sent twice", func() {
				dose := buildAndStoreEntity(0, scheduleStart)
				dose.MarkReminderSent()
				gomega.Expect(doseRepo.Update(ctx, dose)).To(gomega.Succeed())

				entries, err := doseRepo.FindDueForReminder(ctx, from, to, 10)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(entries).To(gomega.BeEmpty())
			})
		})

		ginkgo.When("the dose falls outside the window", func() {
			ginkgo.It("should be excluded", func() {
				buildAndStoreEntity(0, scheduleStart.Add(4*time.Hour))

				entries, err := doseRepo.FindDueForReminder(ctx, from, to, 10)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(entries).To(gomega.BeEmpty())
			})
		})

		ginkgo.When("the treatment is paused", func() {
			ginkgo.It("should be excluded", func() {
				buildAndStoreEntity(0, scheduleStart)
				treatment.Deactivate()
				gomega.Expect(treatmentRepo.Update(ctx, treatment)).To(gomega.Succeed())

				entries, err := doseRepo.FindDueForReminder(ctx, from, to, 10)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(entries).To(gomega.BeEmpty())
			})
		})

		ginkgo.When("the treatment is deleted", func() {
			ginkgo.It("should be excluded", func() {
				buildAndStoreEntity(0, scheduleStart)
				gomega.Expect(treatmentRepo.Delete(ctx, treatment.ID)).To(gomega.Succeed())

				entries, err := doseRepo.FindDueForReminder(ctx, from, to, 10)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(entries).To(gomega.BeEmpty())
			})
		})

		ginkgo.When("the patient is deleted", func() {
			ginkgo.It("should be excluded", func() {
				buildAndStoreEntity(0, scheduleStart)
				gomega.Expect(patientRepo.Delete(ctx, patient.ID)).To(gomega.Succeed())

				entries, err := doseRepo.FindDueForReminder(ctx, from, to, 10)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(entries).To(gomega.BeEmpty())
			})
		})
	})

	ginkgo.Context("FindAgenda", func() {
		ginkgo.When("the tenant has doses in the window", func() {
			ginkgo.It("should return them ordered by schedule", func() {
				buildAndStoreEntity(1, scheduleStart.Add(8*time.Hour))
				buildAndStoreEntity(0, scheduleStart)

				entries, err := doseRepo.FindAgenda(ctx, "tenant-1", scheduleStart.Add(-1*time.Hour), scheduleStart.Add(24*time.Hour), 100)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(entries).To(gomega.HaveLen(2))
				gomega.Expect(entries[0].Dose.SequenceNumber).To(gomega.Equal(0))
				gomega.Expect(entries[1].Dose.SequenceNumber).To(gomega.Equal(1))
			})
		})

		ginkgo.When("the doses belong to another tenant", func() {
			ginkgo.It("should not leak across tenants", func() {
				buildAndStoreEntity(0, scheduleStart)

				entries, err := doseRepo.FindAgenda(ctx, "tenant-2", scheduleStart.Add(-1*time.Hour), scheduleStart.Add(24*time.Hour), 100)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(entries).To(gomega.BeEmpty())
			})
		})
	})

	ginkgo.Context("DeletePendingFrom", func() {
		ginkgo.When("a treatment schedule changes", func() {
			ginkgo.It("should drop future pending doses but keep the record of what happened", func() {
				resolved := buildAndStoreEntity(0, scheduleStart)
				gomega.Expect(resolved.MarkAdministered("user-1", nil)).To(gomega.Succeed())
				gomega.Expect(doseRepo.Update(ctx, resolved)).To(gomega.Succeed())

				pastPending := buildAndStoreEntity(1, time.Now().UTC().Round(0).Add(-2*time.Hour))
				futurePending := buildAndStoreEntity(2, time.Now().UTC().Round(0).Add(6*time.Hour))

				gomega.Expect(doseRepo.DeletePendingFrom(ctx, treatment.ID, time.Now())).To(gomega.Succeed())

				stillThere, err := doseRepo.GetByID(ctx, resolved.ID)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(stillThere.IsDeleted()).To(gomega.BeFalse())

				past, err := doseRepo.GetByID(ctx, pastPending.ID)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(past.IsDeleted()).To(gomega.BeFalse())

				_, err = doseRepo.GetByID(ctx, futurePending.ID)
				gomega.Expect(err).To(gomega.MatchError(medicinesUsecases.ErrDoseNotFound))
			})

			ginkgo.It("should free the sequence numbers so the worker can rebuild them", func() {
				futurePending := buildAndStoreEntity(5, time.Now().UTC().Round(0).Add(6*time.Hour))

				gomega.Expect(doseRepo.DeletePendingFrom(ctx, treatment.ID, time.Now())).To(gomega.Succeed())

				// Re-materialising the same index must not collide with a
				// tombstone: the unique index does not exclude deleted rows, so
				// a soft delete here would break the treatment for good.
				rebuilt, err := medicinesDomain.NewDoseBuilder().
					WithTreatmentID(treatment.ID.String()).
					WithSequenceNumber(futurePending.SequenceNumber).
					WithScheduledAt(time.Now().UTC().Round(0).Add(9*time.Hour)).
					WithDose(15, medicinesDomain.DoseUnitDrop).
					Build()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				gomega.Expect(doseRepo.CreateBatch(ctx, []medicinesDomain.Dose{rebuilt})).To(gomega.Succeed())
			})
		})
	})
})
