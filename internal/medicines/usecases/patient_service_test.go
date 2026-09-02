package usecases_test

import (
	"context"
	"errors"

	medicinesDomain "zensor-server/internal/medicines/domain"
	medicinesUsecases "zensor-server/internal/medicines/usecases"
	shareddomain "zensor-server/internal/shared_kernel/domain"
	sharedUsecases "zensor-server/internal/shared_kernel/usecases"

	mockmedicines "zensor-server/test/unit/doubles/medicines/usecases"
	mocksharedkernel "zensor-server/test/unit/doubles/shared_kernel/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("PatientService", func() {
	var (
		ctrl                *gomock.Controller
		patientRepository   *mockmedicines.MockPatientRepository
		treatmentRepository *mockmedicines.MockTreatmentRepository
		tenantService       *mocksharedkernel.MockTenantService
		service             medicinesUsecases.PatientService
		ctx                 context.Context
		patient             medicinesDomain.Patient
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		patientRepository = mockmedicines.NewMockPatientRepository(ctrl)
		treatmentRepository = mockmedicines.NewMockTreatmentRepository(ctrl)
		tenantService = mocksharedkernel.NewMockTenantService(ctrl)
		service = medicinesUsecases.NewPatientService(patientRepository, treatmentRepository, tenantService)
		ctx = context.Background()

		var err error
		patient, err = medicinesDomain.NewPatientBuilder().
			WithTenantID("tenant-1").
			WithName("Luna").
			WithKind(medicinesDomain.PatientKindAnimal).
			Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Context("CreatePatient", func() {
		ginkgo.When("the tenant exists", func() {
			ginkgo.It("should persist the patient", func() {
				tenantService.EXPECT().GetTenant(ctx, patient.TenantID).Return(shareddomain.Tenant{}, nil)
				patientRepository.EXPECT().Create(ctx, patient).Return(nil)

				gomega.Expect(service.CreatePatient(ctx, patient)).To(gomega.Succeed())
			})
		})

		ginkgo.When("the tenant does not exist", func() {
			ginkgo.It("should fail without touching the repository", func() {
				tenantService.EXPECT().GetTenant(ctx, patient.TenantID).
					Return(shareddomain.Tenant{}, sharedUsecases.ErrTenantNotFound)

				err := service.CreatePatient(ctx, patient)

				gomega.Expect(err).To(gomega.MatchError(sharedUsecases.ErrTenantNotFound))
			})
		})
	})

	ginkgo.Context("GetPatient", func() {
		ginkgo.When("the patient is soft deleted", func() {
			ginkgo.It("should report it as not found", func() {
				patient.SoftDelete()
				patientRepository.EXPECT().GetByID(ctx, patient.ID).Return(patient, nil)

				_, err := service.GetPatient(ctx, patient.ID)

				gomega.Expect(err).To(gomega.MatchError(medicinesUsecases.ErrPatientNotFound))
			})
		})
	})

	ginkgo.Context("DeletePatient", func() {
		ginkgo.When("the patient still has active treatments", func() {
			ginkgo.It("should refuse to delete", func() {
				patientRepository.EXPECT().GetByID(ctx, patient.ID).Return(patient, nil)
				treatmentRepository.EXPECT().CountActiveByPatient(ctx, patient.ID).Return(2, nil)

				err := service.DeletePatient(ctx, patient.ID)

				gomega.Expect(err).To(gomega.MatchError(medicinesUsecases.ErrPatientHasActiveTreatments))
			})
		})

		ginkgo.When("the patient has no active treatments", func() {
			ginkgo.It("should delete it", func() {
				patientRepository.EXPECT().GetByID(ctx, patient.ID).Return(patient, nil)
				treatmentRepository.EXPECT().CountActiveByPatient(ctx, patient.ID).Return(0, nil)
				patientRepository.EXPECT().Delete(ctx, patient.ID).Return(nil)

				gomega.Expect(service.DeletePatient(ctx, patient.ID)).To(gomega.Succeed())
			})
		})

		ginkgo.When("the patient does not exist", func() {
			ginkgo.It("should propagate the error", func() {
				patientRepository.EXPECT().GetByID(ctx, patient.ID).
					Return(medicinesDomain.Patient{}, medicinesUsecases.ErrPatientNotFound)

				err := service.DeletePatient(ctx, patient.ID)

				gomega.Expect(errors.Is(err, medicinesUsecases.ErrPatientNotFound)).To(gomega.BeTrue())
			})
		})
	})
})
