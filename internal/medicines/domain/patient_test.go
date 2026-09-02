package domain_test

import (
	medicinesDomain "zensor-server/internal/medicines/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("Patient", func() {
	ginkgo.Context("Build", func() {
		ginkgo.When("all required values are present", func() {
			ginkgo.It("should build an animal patient", func() {
				patient, err := medicinesDomain.NewPatientBuilder().
					WithTenantID("tenant-1").
					WithName("Luna").
					WithKind(medicinesDomain.PatientKindAnimal).
					WithNotes("perra, 12 kg").
					Build()

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(patient.ID).NotTo(gomega.BeEmpty())
				gomega.Expect(patient.Kind).To(gomega.Equal(medicinesDomain.PatientKindAnimal))
				gomega.Expect(patient.IsDeleted()).To(gomega.BeFalse())
			})
		})

		ginkgo.When("the tenant is missing", func() {
			ginkgo.It("should fail", func() {
				_, err := medicinesDomain.NewPatientBuilder().
					WithName("Luna").
					WithKind(medicinesDomain.PatientKindAnimal).
					Build()
				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrTenantIDRequired))
			})
		})

		ginkgo.When("the name is missing", func() {
			ginkgo.It("should fail", func() {
				_, err := medicinesDomain.NewPatientBuilder().
					WithTenantID("tenant-1").
					WithKind(medicinesDomain.PatientKindHuman).
					Build()
				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrPatientNameRequired))
			})
		})

		ginkgo.When("the kind is unknown", func() {
			ginkgo.It("should fail", func() {
				_, err := medicinesDomain.NewPatientBuilder().
					WithTenantID("tenant-1").
					WithName("Luna").
					WithKind("plant").
					Build()
				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrInvalidPatientKind))
			})
		})
	})

	ginkgo.Context("SoftDelete", func() {
		ginkgo.When("a patient is removed", func() {
			ginkgo.It("should be marked as deleted", func() {
				patient, err := medicinesDomain.NewPatientBuilder().
					WithTenantID("tenant-1").
					WithName("Luna").
					WithKind(medicinesDomain.PatientKindAnimal).
					Build()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				patient.SoftDelete()

				gomega.Expect(patient.IsDeleted()).To(gomega.BeTrue())
			})
		})
	})
})
