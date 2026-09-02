package usecases_test

import (
	"context"
	"time"

	medicinesDomain "zensor-server/internal/medicines/domain"
	medicinesUsecases "zensor-server/internal/medicines/usecases"

	mockmedicines "zensor-server/test/unit/doubles/medicines/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("TreatmentService", func() {
	var (
		ctrl                *gomock.Controller
		treatmentRepository *mockmedicines.MockTreatmentRepository
		patientRepository   *mockmedicines.MockPatientRepository
		doseRepository      *mockmedicines.MockDoseRepository
		service             medicinesUsecases.TreatmentService
		ctx                 context.Context
		patient             medicinesDomain.Patient
		treatment           medicinesDomain.Treatment
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		treatmentRepository = mockmedicines.NewMockTreatmentRepository(ctrl)
		patientRepository = mockmedicines.NewMockPatientRepository(ctrl)
		doseRepository = mockmedicines.NewMockDoseRepository(ctrl)
		service = medicinesUsecases.NewTreatmentService(treatmentRepository, patientRepository, doseRepository)
		ctx = context.Background()

		var err error
		patient, err = medicinesDomain.NewPatientBuilder().
			WithTenantID("tenant-1").
			WithName("Luna").
			WithKind(medicinesDomain.PatientKindAnimal).
			Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		treatment, err = medicinesDomain.NewTreatmentBuilder().
			WithTenantID("tenant-1").
			WithPatientID(patient.ID.String()).
			WithMedicineName("Amoxicilina").
			WithDose(15, medicinesDomain.DoseUnitDrop).
			WithSchedule(medicinesDomain.MedicineSchedule{
				StartAt: time.Now().Add(1 * time.Hour),
				Every:   8,
				Unit:    medicinesDomain.IntervalUnitHour,
			}).
			Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Context("CreateTreatment", func() {
		ginkgo.When("the patient exists in the same tenant", func() {
			ginkgo.It("should persist the treatment", func() {
				patientRepository.EXPECT().GetByID(ctx, treatment.PatientID).Return(patient, nil)
				treatmentRepository.EXPECT().Create(ctx, treatment).Return(nil)

				gomega.Expect(service.CreateTreatment(ctx, treatment)).To(gomega.Succeed())
			})
		})

		ginkgo.When("the patient is deleted", func() {
			ginkgo.It("should refuse", func() {
				patient.SoftDelete()
				patientRepository.EXPECT().GetByID(ctx, treatment.PatientID).Return(patient, nil)

				err := service.CreateTreatment(ctx, treatment)

				gomega.Expect(err).To(gomega.MatchError(medicinesUsecases.ErrPatientDeleted))
			})
		})

		ginkgo.When("the patient belongs to another tenant", func() {
			ginkgo.It("should refuse", func() {
				patient.TenantID = "tenant-2"
				patientRepository.EXPECT().GetByID(ctx, treatment.PatientID).Return(patient, nil)

				err := service.CreateTreatment(ctx, treatment)

				gomega.Expect(err).To(gomega.MatchError(medicinesUsecases.ErrTenantMismatch))
			})
		})

		ginkgo.When("the schedule cannot produce doses", func() {
			ginkgo.It("should refuse before persisting, so the worker never sees it", func() {
				treatment.Schedule.Every = 0
				patientRepository.EXPECT().GetByID(ctx, treatment.PatientID).Return(patient, nil)

				err := service.CreateTreatment(ctx, treatment)

				gomega.Expect(err).To(gomega.MatchError(medicinesUsecases.ErrInvalidTreatmentSchedule))
				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrIntervalRequired))
			})
		})
	})

	ginkgo.Context("UpdateTreatment", func() {
		ginkgo.When("the schedule changes", func() {
			ginkgo.It("should discard the pending doses built from the old schedule", func() {
				treatmentRepository.EXPECT().GetByID(ctx, treatment.ID).Return(treatment, nil)
				treatmentRepository.EXPECT().Update(ctx, treatment).Return(nil)
				doseRepository.EXPECT().
					DeletePendingFrom(ctx, treatment.ID, gomock.Any()).
					Return(nil)

				gomega.Expect(service.UpdateTreatment(ctx, treatment)).To(gomega.Succeed())
			})
		})

		ginkgo.When("the treatment is deleted", func() {
			ginkgo.It("should refuse", func() {
				stored := treatment
				stored.SoftDelete()
				treatmentRepository.EXPECT().GetByID(ctx, treatment.ID).Return(stored, nil)

				err := service.UpdateTreatment(ctx, treatment)

				gomega.Expect(err).To(gomega.MatchError(medicinesUsecases.ErrTreatmentDeleted))
			})
		})
	})

	ginkgo.Context("DeleteTreatment", func() {
		ginkgo.When("the treatment exists", func() {
			ginkgo.It("should delete it and discard its pending doses", func() {
				treatmentRepository.EXPECT().GetByID(ctx, treatment.ID).Return(treatment, nil)
				treatmentRepository.EXPECT().Delete(ctx, treatment.ID).Return(nil)
				doseRepository.EXPECT().DeletePendingFrom(ctx, treatment.ID, gomock.Any()).Return(nil)

				gomega.Expect(service.DeleteTreatment(ctx, treatment.ID)).To(gomega.Succeed())
			})
		})
	})

	ginkgo.Context("DeactivateTreatment", func() {
		ginkgo.When("the treatment is paused", func() {
			ginkgo.It("should discard its pending doses", func() {
				treatmentRepository.EXPECT().GetByID(ctx, treatment.ID).Return(treatment, nil)
				treatmentRepository.EXPECT().
					Update(ctx, gomock.Cond(func(x any) bool {
						t, ok := x.(medicinesDomain.Treatment)
						return ok && !t.IsActive
					})).
					Return(nil)
				doseRepository.EXPECT().DeletePendingFrom(ctx, treatment.ID, gomock.Any()).Return(nil)

				gomega.Expect(service.DeactivateTreatment(ctx, treatment.ID)).To(gomega.Succeed())
			})
		})
	})

	ginkgo.Context("ActivateTreatment", func() {
		ginkgo.When("the treatment is resumed", func() {
			ginkgo.It("should not discard doses, the schedule has not changed", func() {
				treatment.Deactivate()
				treatmentRepository.EXPECT().GetByID(ctx, treatment.ID).Return(treatment, nil)
				treatmentRepository.EXPECT().
					Update(ctx, gomock.Cond(func(x any) bool {
						t, ok := x.(medicinesDomain.Treatment)
						return ok && t.IsActive
					})).
					Return(nil)

				gomega.Expect(service.ActivateTreatment(ctx, treatment.ID)).To(gomega.Succeed())
			})
		})
	})
})
