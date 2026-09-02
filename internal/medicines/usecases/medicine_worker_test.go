package usecases_test

import (
	"context"
	"errors"
	"time"
	"zensor-server/internal/infra/async"

	medicinesDomain "zensor-server/internal/medicines/domain"
	medicinesUsecases "zensor-server/internal/medicines/usecases"
	shareddomain "zensor-server/internal/shared_kernel/domain"

	mockasync "zensor-server/test/unit/doubles/infra/async"
	mockmedicines "zensor-server/test/unit/doubles/medicines/usecases"
	mocksharedkernel "zensor-server/test/unit/doubles/shared_kernel/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("MedicineWorker", func() {
	var (
		ctrl                *gomock.Controller
		treatmentRepository *mockmedicines.MockTreatmentRepository
		doseRepository      *mockmedicines.MockDoseRepository
		tenantService       *mocksharedkernel.MockTenantService
		tenantConfigService *mocksharedkernel.MockTenantConfigurationService
		broker              *mockasync.MockInternalBroker
		worker              *medicinesUsecases.MedicineWorker
		ctx                 context.Context
		patient             medicinesDomain.Patient
		treatment           medicinesDomain.Treatment
	)

	newTreatment := func(every int, unit medicinesDomain.IntervalUnit, startAt time.Time) medicinesDomain.Treatment {
		built, err := medicinesDomain.NewTreatmentBuilder().
			WithTenantID("tenant-1").
			WithPatientID(patient.ID.String()).
			WithMedicineName("Amoxicilina").
			WithDose(15, medicinesDomain.DoseUnitDrop).
			WithSchedule(medicinesDomain.MedicineSchedule{StartAt: startAt, Every: every, Unit: unit}).
			Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
		return built
	}

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		treatmentRepository = mockmedicines.NewMockTreatmentRepository(ctrl)
		doseRepository = mockmedicines.NewMockDoseRepository(ctrl)
		tenantService = mocksharedkernel.NewMockTenantService(ctrl)
		tenantConfigService = mocksharedkernel.NewMockTenantConfigurationService(ctrl)
		broker = mockasync.NewMockInternalBroker(ctrl)
		worker = medicinesUsecases.NewMedicineWorker(
			time.NewTicker(1*time.Hour),
			treatmentRepository,
			doseRepository,
			tenantService,
			tenantConfigService,
			broker,
		)
		ctx = context.Background()

		var err error
		patient, err = medicinesDomain.NewPatientBuilder().
			WithTenantID("tenant-1").
			WithName("Luna").
			WithKind(medicinesDomain.PatientKindAnimal).
			Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		treatment = newTreatment(8, medicinesDomain.IntervalUnitHour, time.Now().Add(-1*time.Hour))
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Context("MaterialiseDoses", func() {
		ginkgo.When("no dose exists yet", func() {
			ginkgo.It("should create every occurrence in the window", func() {
				var created []medicinesDomain.Dose
				treatmentRepository.EXPECT().FindAllActive(ctx).Return([]medicinesDomain.Treatment{treatment}, nil)
				doseRepository.EXPECT().
					FindByTreatmentInWindow(ctx, treatment.ID, gomock.Any(), gomock.Any()).
					Return(nil, nil)
				doseRepository.EXPECT().
					CreateBatch(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, doses []medicinesDomain.Dose) error {
						created = doses
						return nil
					})

				worker.MaterialiseDoses(ctx)

				gomega.Expect(created).NotTo(gomega.BeEmpty())
				gomega.Expect(created[0].Status).To(gomega.Equal(medicinesDomain.DoseStatusPending))
				gomega.Expect(created[0].Quantity).To(gomega.BeEquivalentTo(15))
			})
		})

		ginkgo.When("the same window is materialised twice", func() {
			ginkgo.It("should be idempotent, recomputing the very same sequence numbers", func() {
				var firstRun []medicinesDomain.Dose

				treatmentRepository.EXPECT().FindAllActive(ctx).Return([]medicinesDomain.Treatment{treatment}, nil)
				doseRepository.EXPECT().
					FindByTreatmentInWindow(ctx, treatment.ID, gomock.Any(), gomock.Any()).
					Return(nil, nil)
				doseRepository.EXPECT().
					CreateBatch(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, doses []medicinesDomain.Dose) error {
						firstRun = doses
						return nil
					})

				worker.MaterialiseDoses(ctx)
				gomega.Expect(firstRun).NotTo(gomega.BeEmpty())

				// Second tick: the doses of the first run already exist.
				treatmentRepository.EXPECT().FindAllActive(ctx).Return([]medicinesDomain.Treatment{treatment}, nil)
				doseRepository.EXPECT().
					FindByTreatmentInWindow(ctx, treatment.ID, gomock.Any(), gomock.Any()).
					Return(firstRun, nil)
				// No CreateBatch expectation: calling it would fail the spec.

				worker.MaterialiseDoses(ctx)
			})
		})

		ginkgo.When("the treatment has a total dose count", func() {
			ginkgo.It("should never create more doses than prescribed", func() {
				total := 3
				treatment.TotalDoses = &total
				var created []medicinesDomain.Dose

				treatmentRepository.EXPECT().FindAllActive(ctx).Return([]medicinesDomain.Treatment{treatment}, nil)
				doseRepository.EXPECT().
					FindByTreatmentInWindow(ctx, treatment.ID, gomock.Any(), gomock.Any()).
					Return(nil, nil)
				doseRepository.EXPECT().
					CreateBatch(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, doses []medicinesDomain.Dose) error {
						created = doses
						return nil
					})

				worker.MaterialiseDoses(ctx)

				gomega.Expect(created).To(gomega.HaveLen(3))
				gomega.Expect(created[2].SequenceNumber).To(gomega.Equal(2))
			})
		})

		ginkgo.When("the treatment has an end date", func() {
			ginkgo.It("should stop at the end date", func() {
				endAt := treatment.Schedule.StartAt.Add(17 * time.Hour)
				treatment.EndAt = &endAt
				var created []medicinesDomain.Dose

				treatmentRepository.EXPECT().FindAllActive(ctx).Return([]medicinesDomain.Treatment{treatment}, nil)
				doseRepository.EXPECT().
					FindByTreatmentInWindow(ctx, treatment.ID, gomock.Any(), gomock.Any()).
					Return(nil, nil)
				doseRepository.EXPECT().
					CreateBatch(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, doses []medicinesDomain.Dose) error {
						created = doses
						return nil
					})

				worker.MaterialiseDoses(ctx)

				gomega.Expect(created).To(gomega.HaveLen(3))
			})
		})

		ginkgo.When("the treatment started while the server was down", func() {
			ginkgo.It("should backfill the doses that already fell due", func() {
				treatment = newTreatment(4, medicinesDomain.IntervalUnitHour, time.Now().Add(-12*time.Hour))
				var created []medicinesDomain.Dose

				treatmentRepository.EXPECT().FindAllActive(ctx).Return([]medicinesDomain.Treatment{treatment}, nil)
				doseRepository.EXPECT().
					FindByTreatmentInWindow(ctx, treatment.ID, gomock.Any(), gomock.Any()).
					Return(nil, nil)
				doseRepository.EXPECT().
					CreateBatch(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, doses []medicinesDomain.Dose) error {
						created = doses
						return nil
					})

				worker.MaterialiseDoses(ctx)

				var pastDoses int
				for _, dose := range created {
					if dose.ScheduledAt.Before(time.Now()) {
						pastDoses++
					}
				}
				gomega.Expect(pastDoses).To(gomega.BeNumerically(">=", 3))
			})
		})

		ginkgo.When("a treatment cannot be materialised", func() {
			ginkgo.It("should keep going with the rest", func() {
				broken := treatment
				broken.Schedule.Every = 0
				healthy := newTreatment(8, medicinesDomain.IntervalUnitHour, time.Now())

				treatmentRepository.EXPECT().FindAllActive(ctx).
					Return([]medicinesDomain.Treatment{broken, healthy}, nil)
				doseRepository.EXPECT().
					FindByTreatmentInWindow(ctx, healthy.ID, gomock.Any(), gomock.Any()).
					Return(nil, nil)
				doseRepository.EXPECT().CreateBatch(ctx, gomock.Any()).Return(nil)

				worker.MaterialiseDoses(ctx)
			})
		})
	})

	ginkgo.Context("EmitReminders", func() {
		var entry medicinesUsecases.DoseWithContext

		ginkgo.BeforeEach(func() {
			dose, err := medicinesDomain.NewDoseBuilder().
				WithTreatmentID(treatment.ID.String()).
				WithSequenceNumber(0).
				WithScheduledAt(time.Now()).
				WithDose(15, medicinesDomain.DoseUnitDrop).
				Build()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			entry = medicinesUsecases.DoseWithContext{
				Dose:      dose,
				Treatment: treatment,
				Patient:   patient,
			}
		})

		expectTenantTimezone := func(timezone string) {
			tenantService.EXPECT().GetTenant(ctx, shareddomain.ID("tenant-1")).
				Return(shareddomain.Tenant{}, nil)
			tenantConfigService.EXPECT().
				GetOrCreateTenantConfiguration(ctx, gomock.Any(), gomock.Any()).
				Return(shareddomain.TenantConfiguration{Timezone: timezone}, nil)
		}

		ginkgo.When("a dose falls due", func() {
			ginkgo.It("should mark the reminder as sent before publishing", func() {
				doseRepository.EXPECT().
					FindDueForReminder(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]medicinesUsecases.DoseWithContext{entry}, nil)
				expectTenantTimezone("UTC")

				gomock.InOrder(
					doseRepository.EXPECT().
						Update(ctx, gomock.Cond(func(x any) bool {
							dose, ok := x.(medicinesDomain.Dose)
							return ok && dose.HasReminderBeenSent()
						})).
						Return(nil),
					broker.EXPECT().
						Publish(ctx, async.BrokerTopicName("medicine_doses"), gomock.Any()).
						Return(nil),
				)

				worker.EmitReminders(ctx)
			})
		})

		ginkgo.When("marking the reminder fails", func() {
			ginkgo.It("should not publish, so a reminder is never sent unmarked", func() {
				doseRepository.EXPECT().
					FindDueForReminder(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]medicinesUsecases.DoseWithContext{entry}, nil)
				doseRepository.EXPECT().Update(ctx, gomock.Any()).Return(errors.New("database is down"))

				worker.EmitReminders(ctx)
			})
		})

		ginkgo.When("a dose due event is published", func() {
			ginkgo.It("should carry exactly the keys the push notification config interpolates", func() {
				var published async.BrokerMessage

				doseRepository.EXPECT().
					FindDueForReminder(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]medicinesUsecases.DoseWithContext{entry}, nil)
				expectTenantTimezone("UTC")
				doseRepository.EXPECT().Update(ctx, gomock.Any()).Return(nil)
				broker.EXPECT().
					Publish(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, _ async.BrokerTopicName, msg async.BrokerMessage) error {
						published = msg
						return nil
					})

				worker.EmitReminders(ctx)

				gomega.Expect(published.Event).To(gomega.Equal("dose_due"))
				payload, ok := published.Value.(map[string]any)
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(payload).To(gomega.HaveLen(8))
				gomega.Expect(payload).To(gomega.HaveKey("dose_id"))
				gomega.Expect(payload).To(gomega.HaveKey("treatment_id"))
				gomega.Expect(payload).To(gomega.HaveKey("patient_id"))
				gomega.Expect(payload).To(gomega.HaveKey("tenant_id"))
				gomega.Expect(payload["patient_name"]).To(gomega.Equal("Luna"))
				gomega.Expect(payload["medicine_name"]).To(gomega.Equal("Amoxicilina"))
				gomega.Expect(payload["dose_text"]).To(gomega.Equal("15 gotas"))
				gomega.Expect(payload["scheduled_time"]).To(gomega.MatchRegexp(`^\d{2}:\d{2}$`))
			})
		})

		ginkgo.When("several doses of the same tenant fall due", func() {
			ginkgo.It("should resolve the tenant timezone once", func() {
				second := entry
				second.Dose.ID = "dose-2"

				doseRepository.EXPECT().
					FindDueForReminder(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]medicinesUsecases.DoseWithContext{entry, second}, nil)
				expectTenantTimezone("America/Santiago")
				doseRepository.EXPECT().Update(ctx, gomock.Any()).Return(nil).Times(2)
				broker.EXPECT().Publish(ctx, gomock.Any(), gomock.Any()).Return(nil).Times(2)

				worker.EmitReminders(ctx)
			})
		})

		ginkgo.When("the tenant timezone cannot be resolved", func() {
			ginkgo.It("should still send the reminder, falling back to UTC", func() {
				doseRepository.EXPECT().
					FindDueForReminder(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]medicinesUsecases.DoseWithContext{entry}, nil)
				tenantService.EXPECT().GetTenant(ctx, gomock.Any()).
					Return(shareddomain.Tenant{}, errors.New("tenant service is down"))
				doseRepository.EXPECT().Update(ctx, gomock.Any()).Return(nil)
				broker.EXPECT().Publish(ctx, gomock.Any(), gomock.Any()).Return(nil)

				worker.EmitReminders(ctx)
			})
		})

		ginkgo.When("nothing is due", func() {
			ginkgo.It("should publish nothing", func() {
				doseRepository.EXPECT().
					FindDueForReminder(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, nil)

				worker.EmitReminders(ctx)
			})
		})
	})
})
