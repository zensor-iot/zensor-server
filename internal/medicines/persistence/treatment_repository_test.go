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

var _ = ginkgo.Describe("TreatmentRepository", func() {
	var (
		orm           sql.ORM
		treatmentRepo medicinesUsecases.TreatmentRepository
		patientRepo   medicinesUsecases.PatientRepository
		ctx           context.Context
		schedule      medicinesDomain.MedicineSchedule
	)

	newTreatment := func(patientID string) medicinesDomain.Treatment {
		treatment, err := medicinesDomain.NewTreatmentBuilder().
			WithTenantID("tenant-1").
			WithPatientID(patientID).
			WithMedicineName("Amoxicilina").
			WithDose(1, medicinesDomain.DoseUnitPill).
			WithSchedule(schedule).
			Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return treatment
	}

	newPatient := func(name string) medicinesDomain.Patient {
		patient, err := medicinesDomain.NewPatientBuilder().
			WithTenantID("tenant-1").
			WithName(name).
			WithKind(medicinesDomain.PatientKindHuman).
			Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		gomega.Expect(patientRepo.Create(ctx, patient)).To(gomega.Succeed())
		return patient
	}

	ginkgo.BeforeEach(func() {
		var err error
		orm, err = sql.NewMemoryORM("migrations")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		patientRepo, err = medicinesPersistence.NewPatientRepository(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		treatmentRepo, err = medicinesPersistence.NewTreatmentRepository(orm)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		ctx = context.Background()
		schedule = medicinesDomain.MedicineSchedule{
			StartAt: time.Now().UTC().Round(0).Add(1 * time.Hour),
			Every:   8,
			Unit:    medicinesDomain.IntervalUnitHour,
		}

		for _, model := range []any{
			&medicinesPersistenceInternal.Treatment{},
			&medicinesPersistenceInternal.Patient{},
		} {
			gomega.Expect(orm.WithContext(ctx).Unscoped().Where("1 = 1").Delete(model).Error()).To(gomega.Succeed())
		}
	})

	ginkgo.Context("Create and GetByID", func() {
		ginkgo.When("a treatment is stored", func() {
			ginkgo.It("should round trip the schedule through its JSON column", func() {
				patient := newPatient("Abuela")
				treatment := newTreatment(patient.ID.String())
				gomega.Expect(treatmentRepo.Create(ctx, treatment)).To(gomega.Succeed())

				stored, err := treatmentRepo.GetByID(ctx, treatment.ID)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(stored.Schedule.Every).To(gomega.Equal(8))
				gomega.Expect(stored.Schedule.Unit).To(gomega.Equal(medicinesDomain.IntervalUnitHour))
				gomega.Expect(stored.Schedule.StartAt.UTC()).To(gomega.BeTemporally("~", schedule.StartAt, time.Second))
				gomega.Expect(stored.DoseText()).To(gomega.Equal("1 pastilla"))
			})
		})

		ginkgo.When("the treatment does not exist", func() {
			ginkgo.It("should report it as not found", func() {
				_, err := treatmentRepo.GetByID(ctx, "missing")
				gomega.Expect(err).To(gomega.MatchError(medicinesUsecases.ErrTreatmentNotFound))
			})
		})

		ginkgo.When("the treatment carries both end conditions", func() {
			ginkgo.It("should round trip them", func() {
				patient := newPatient("Abuela")
				treatment := newTreatment(patient.ID.String())
				total := 21
				endAt := schedule.StartAt.Add(240 * time.Hour)
				treatment.TotalDoses = &total
				treatment.EndAt = &endAt
				gomega.Expect(treatmentRepo.Create(ctx, treatment)).To(gomega.Succeed())

				stored, err := treatmentRepo.GetByID(ctx, treatment.ID)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(*stored.TotalDoses).To(gomega.Equal(21))
				gomega.Expect(stored.EndAt).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Context("CountActiveByPatient", func() {
		ginkgo.When("the patient has active and paused treatments", func() {
			ginkgo.It("should count only the active ones", func() {
				patient := newPatient("Abuela")

				active := newTreatment(patient.ID.String())
				gomega.Expect(treatmentRepo.Create(ctx, active)).To(gomega.Succeed())

				paused := newTreatment(patient.ID.String())
				paused.Deactivate()
				gomega.Expect(treatmentRepo.Create(ctx, paused)).To(gomega.Succeed())

				count, err := treatmentRepo.CountActiveByPatient(ctx, patient.ID)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(count).To(gomega.Equal(1))
			})
		})

		ginkgo.When("a treatment is deleted", func() {
			ginkgo.It("should stop counting it", func() {
				patient := newPatient("Abuela")
				treatment := newTreatment(patient.ID.String())
				gomega.Expect(treatmentRepo.Create(ctx, treatment)).To(gomega.Succeed())
				gomega.Expect(treatmentRepo.Delete(ctx, treatment.ID)).To(gomega.Succeed())

				count, err := treatmentRepo.CountActiveByPatient(ctx, patient.ID)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(count).To(gomega.BeZero())
			})
		})
	})

	ginkgo.Context("FindAllActive", func() {
		ginkgo.When("some treatments are paused or deleted", func() {
			ginkgo.It("should return only the ones the worker should act on", func() {
				patient := newPatient("Abuela")

				active := newTreatment(patient.ID.String())
				gomega.Expect(treatmentRepo.Create(ctx, active)).To(gomega.Succeed())

				paused := newTreatment(patient.ID.String())
				paused.Deactivate()
				gomega.Expect(treatmentRepo.Create(ctx, paused)).To(gomega.Succeed())

				deleted := newTreatment(patient.ID.String())
				gomega.Expect(treatmentRepo.Create(ctx, deleted)).To(gomega.Succeed())
				gomega.Expect(treatmentRepo.Delete(ctx, deleted.ID)).To(gomega.Succeed())

				treatments, err := treatmentRepo.FindAllActive(ctx)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(treatments).To(gomega.HaveLen(1))
				gomega.Expect(treatments[0].ID).To(gomega.Equal(active.ID))
			})
		})
	})

	ginkgo.Context("FindAllByPatient", func() {
		ginkgo.When("treatments belong to different patients", func() {
			ginkgo.It("should return only the ones of the requested patient", func() {
				first := newPatient("Abuela")
				second := newPatient("Luna")

				gomega.Expect(treatmentRepo.Create(ctx, newTreatment(first.ID.String()))).To(gomega.Succeed())
				gomega.Expect(treatmentRepo.Create(ctx, newTreatment(second.ID.String()))).To(gomega.Succeed())

				treatments, total, err := treatmentRepo.FindAllByPatient(ctx, first.ID, medicinesUsecases.Pagination{Limit: 10})

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(total).To(gomega.Equal(1))
				gomega.Expect(treatments).To(gomega.HaveLen(1))
			})
		})
	})
})
