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

var _ = ginkgo.Describe("DoseService", func() {
	var (
		ctrl           *gomock.Controller
		doseRepository *mockmedicines.MockDoseRepository
		service        medicinesUsecases.DoseService
		ctx            context.Context
		dose           medicinesDomain.Dose
	)

	newDose := func(scheduledAt time.Time) medicinesDomain.Dose {
		built, err := medicinesDomain.NewDoseBuilder().
			WithTreatmentID("treatment-1").
			WithSequenceNumber(0).
			WithScheduledAt(scheduledAt).
			WithDose(15, medicinesDomain.DoseUnitDrop).
			Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return built
	}

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		doseRepository = mockmedicines.NewMockDoseRepository(ctrl)
		service = medicinesUsecases.NewDoseService(doseRepository)
		ctx = context.Background()
		dose = newDose(time.Now().Add(-10 * time.Minute))
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Context("AdministerDose", func() {
		ginkgo.When("the dose is pending", func() {
			ginkgo.It("should record it as administered", func() {
				doseRepository.EXPECT().GetByID(ctx, dose.ID).Return(dose, nil)
				doseRepository.EXPECT().
					Update(ctx, gomock.Cond(func(x any) bool {
						d, ok := x.(medicinesDomain.Dose)
						return ok && d.Status == medicinesDomain.DoseStatusAdministered &&
							d.ResolvedBy != nil && *d.ResolvedBy == "user-1"
					})).
					Return(nil)

				gomega.Expect(service.AdministerDose(ctx, dose.ID, "user-1", nil)).To(gomega.Succeed())
			})
		})

		ginkgo.When("the dose was already resolved", func() {
			ginkgo.It("should refuse without updating", func() {
				gomega.Expect(dose.MarkSkipped("user-1", nil)).To(gomega.Succeed())
				doseRepository.EXPECT().GetByID(ctx, dose.ID).Return(dose, nil)

				err := service.AdministerDose(ctx, dose.ID, "user-2", nil)

				gomega.Expect(err).To(gomega.MatchError(medicinesDomain.ErrDoseAlreadyResolved))
			})
		})

		ginkgo.When("the dose is scheduled a little ahead", func() {
			ginkgo.It("should still allow it, giving medicine early is normal care", func() {
				future := newDose(time.Now().Add(30 * time.Minute))
				doseRepository.EXPECT().GetByID(ctx, future.ID).Return(future, nil)
				doseRepository.EXPECT().Update(ctx, gomock.Any()).Return(nil)

				gomega.Expect(service.AdministerDose(ctx, future.ID, "user-1", nil)).To(gomega.Succeed())
			})
		})

		ginkgo.When("the dose is deleted", func() {
			ginkgo.It("should report it as not found", func() {
				dose.SoftDelete()
				doseRepository.EXPECT().GetByID(ctx, dose.ID).Return(dose, nil)

				err := service.AdministerDose(ctx, dose.ID, "user-1", nil)

				gomega.Expect(err).To(gomega.MatchError(medicinesUsecases.ErrDoseNotFound))
			})
		})
	})

	ginkgo.Context("SkipDose", func() {
		ginkgo.When("the dose is pending", func() {
			ginkgo.It("should record it as skipped with its notes", func() {
				notes := medicinesDomain.DoseNotes("la escupió")
				doseRepository.EXPECT().GetByID(ctx, dose.ID).Return(dose, nil)
				doseRepository.EXPECT().
					Update(ctx, gomock.Cond(func(x any) bool {
						d, ok := x.(medicinesDomain.Dose)
						return ok && d.Status == medicinesDomain.DoseStatusSkipped &&
							d.Notes != nil && *d.Notes == notes
					})).
					Return(nil)

				gomega.Expect(service.SkipDose(ctx, dose.ID, "user-1", &notes)).To(gomega.Succeed())
			})
		})
	})

	ginkgo.Context("ListAgenda", func() {
		ginkgo.When("the window is within the limit", func() {
			ginkgo.It("should return the entries", func() {
				from := time.Now()
				to := from.Add(24 * time.Hour)
				doseRepository.EXPECT().
					FindAgenda(ctx, gomock.Any(), from, to, gomock.Any()).
					Return([]medicinesUsecases.DoseWithContext{{Dose: dose}}, nil)

				entries, err := service.ListAgenda(ctx, "tenant-1", from, to)

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(entries).To(gomega.HaveLen(1))
			})
		})

		ginkgo.When("the window is longer than a month", func() {
			ginkgo.It("should refuse without querying", func() {
				from := time.Now()
				to := from.Add(40 * 24 * time.Hour)

				_, err := service.ListAgenda(ctx, "tenant-1", from, to)

				gomega.Expect(err).To(gomega.MatchError(medicinesUsecases.ErrAgendaWindowTooLarge))
			})
		})
	})
})
