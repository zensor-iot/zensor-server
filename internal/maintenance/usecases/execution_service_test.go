package usecases_test

import (
	"context"
	"errors"
	"time"
	"zensor-server/internal/infra/utils"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	maintenanceUsecases "zensor-server/internal/maintenance/usecases"
	shareddomain "zensor-server/internal/shared_kernel/domain"
	mockmaintenance "zensor-server/test/unit/doubles/maintenance/usecases"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("MaintenanceExecutionService", func() {
	var (
		ctrl                   *gomock.Controller
		mockRepository         *mockmaintenance.MockExecutionRepository
		mockActivityRepository *mockmaintenance.MockActivityRepository
		service                maintenanceUsecases.ExecutionService
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockRepository = mockmaintenance.NewMockExecutionRepository(ctrl)
		mockActivityRepository = mockmaintenance.NewMockActivityRepository(ctrl)
		service = maintenanceUsecases.NewExecutionService(mockRepository, mockActivityRepository)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("CreateExecution", func() {
		var execution maintenanceDomain.Execution
		var activity maintenanceDomain.Activity

		BeforeEach(func() {
			activityID := shareddomain.ID(utils.GenerateUUID())
			activityType := maintenanceDomain.ActivityType{
				Name:         shareddomain.Name(maintenanceDomain.ActivityTypeWaterSystem),
				DisplayName:  shareddomain.DisplayName("Water System"),
				Description:  shareddomain.Description("Water system maintenance"),
				IsPredefined: true,
				Fields:       []maintenanceDomain.FieldDefinition{},
			}

			activity, _ = maintenanceDomain.NewActivityBuilder().
				WithTenantID(shareddomain.ID(utils.GenerateUUID())).
				WithType(activityType).
				WithName("Test Activity").
				WithDescription("Test Description").
				WithSchedule("0 0 1 * *").
				WithFields([]maintenanceDomain.FieldDefinition{}).
				Build()
			activity.ID = activityID

			execution, _ = maintenanceDomain.NewExecutionBuilder().
				WithActivityID(activityID).
				WithScheduledDate(time.Now().AddDate(0, 0, 30)).
				WithFieldValues(map[string]any{"maintenance_type": "Filter Replacement"}).
				Build()
		})

		When("creating a valid execution", func() {
			It("should successfully create the execution", func() {
				mockActivityRepository.EXPECT().
					GetByID(gomock.Any(), execution.ActivityID).
					Return(activity, nil)
				mockRepository.EXPECT().
					Create(gomock.Any(), execution).
					Return(nil)

				err := service.CreateExecution(context.Background(), execution)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("activity does not exist", func() {
			It("should return an error", func() {
				mockActivityRepository.EXPECT().
					GetByID(gomock.Any(), execution.ActivityID).
					Return(maintenanceDomain.Activity{}, maintenanceUsecases.ErrActivityNotFound)

				err := service.CreateExecution(context.Background(), execution)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("activity not found"))
			})
		})

		When("repository returns an error", func() {
			It("should return the error", func() {
				mockActivityRepository.EXPECT().
					GetByID(gomock.Any(), execution.ActivityID).
					Return(activity, nil)
				mockRepository.EXPECT().
					Create(gomock.Any(), execution).
					Return(errors.New("database error"))

				err := service.CreateExecution(context.Background(), execution)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("creating maintenance execution"))
			})
		})
	})

	Context("GetExecution", func() {
		var executionID shareddomain.ID
		var execution maintenanceDomain.Execution

		BeforeEach(func() {
			activityID := shareddomain.ID(utils.GenerateUUID())
			executionID = shareddomain.ID(utils.GenerateUUID())

			execution, _ = maintenanceDomain.NewExecutionBuilder().
				WithActivityID(activityID).
				WithScheduledDate(time.Now().AddDate(0, 0, 30)).
				WithFieldValues(map[string]any{"maintenance_type": "Filter Replacement"}).
				Build()
			execution.ID = executionID
		})

		When("execution exists", func() {
			It("should return the execution", func() {
				mockRepository.EXPECT().
					GetByID(gomock.Any(), executionID).
					Return(execution, nil)

				result, err := service.GetExecution(context.Background(), executionID)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ID).To(Equal(executionID))
				Expect(result.ActivityID).To(Equal(execution.ActivityID))
			})
		})

		When("execution does not exist", func() {
			It("should return ErrMaintenanceExecutionNotFound", func() {
				mockRepository.EXPECT().
					GetByID(gomock.Any(), executionID).
					Return(maintenanceDomain.Execution{}, maintenanceUsecases.ErrExecutionNotFound)

				result, err := service.GetExecution(context.Background(), executionID)
				Expect(err).To(MatchError(maintenanceUsecases.ErrExecutionNotFound))
				Expect(result.ID).To(BeEmpty())
			})
		})

		When("repository returns an error", func() {
			It("should return the error", func() {
				mockRepository.EXPECT().
					GetByID(gomock.Any(), executionID).
					Return(maintenanceDomain.Execution{}, errors.New("database error"))

				result, err := service.GetExecution(context.Background(), executionID)
				Expect(err).To(HaveOccurred())
				Expect(result.ID).To(BeEmpty())
				Expect(err.Error()).To(ContainSubstring("getting maintenance execution"))
			})
		})
	})

	Context("ListExecutionsByActivity", func() {
		var activityID shareddomain.ID
		var executions []maintenanceDomain.Execution

		BeforeEach(func() {
			activityID = shareddomain.ID(utils.GenerateUUID())
			executions = []maintenanceDomain.Execution{}

			for i := 0; i < 3; i++ {
				execution, _ := maintenanceDomain.NewExecutionBuilder().
					WithActivityID(activityID).
					WithScheduledDate(time.Now().AddDate(0, 0, 30+i*30)).
					WithFieldValues(map[string]any{"maintenance_type": "Filter Replacement"}).
					Build()
				executions = append(executions, execution)
			}
		})

		When("listing executions", func() {
			It("should return all executions for the activity", func() {
				pagination := maintenanceUsecases.Pagination{Limit: 10, Offset: 0}
				mockRepository.EXPECT().
					FindAllByActivity(gomock.Any(), activityID, pagination).
					Return(executions, len(executions), nil)

				result, total, err := service.ListExecutionsByActivity(context.Background(), activityID, pagination)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(3))
				Expect(total).To(Equal(3))
			})
		})

		When("repository returns an error", func() {
			It("should return the error", func() {
				pagination := maintenanceUsecases.Pagination{Limit: 10, Offset: 0}
				mockRepository.EXPECT().
					FindAllByActivity(gomock.Any(), activityID, pagination).
					Return(nil, 0, errors.New("database error"))

				result, total, err := service.ListExecutionsByActivity(context.Background(), activityID, pagination)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
				Expect(total).To(Equal(0))
				Expect(err.Error()).To(ContainSubstring("listing maintenance executions"))
			})
		})
	})

	Context("MarkExecutionCompleted", func() {
		var execution maintenanceDomain.Execution
		var completedBy string

		BeforeEach(func() {
			activityID := shareddomain.ID(utils.GenerateUUID())
			executionID := shareddomain.ID(utils.GenerateUUID())
			completedBy = "user@example.com"

			execution, _ = maintenanceDomain.NewExecutionBuilder().
				WithActivityID(activityID).
				WithScheduledDate(time.Now().AddDate(0, 0, -5)).
				WithFieldValues(map[string]any{"maintenance_type": "Filter Replacement"}).
				Build()
			execution.ID = executionID
		})

		When("marking completion for an existing execution", func() {
			It("should successfully mark the execution as completed", func() {
				mockRepository.EXPECT().
					GetByID(gomock.Any(), execution.ID).
					Return(execution, nil)
				mockRepository.EXPECT().
					MarkCompleted(gomock.Any(), execution.ID, completedBy).
					Return(nil)

				err := service.MarkExecutionCompleted(context.Background(), execution.ID, completedBy)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		When("execution does not exist", func() {
			It("should return ErrMaintenanceExecutionNotFound", func() {
				mockRepository.EXPECT().
					GetByID(gomock.Any(), execution.ID).
					Return(maintenanceDomain.Execution{}, maintenanceUsecases.ErrExecutionNotFound)

				err := service.MarkExecutionCompleted(context.Background(), execution.ID, completedBy)
				Expect(err).To(MatchError(maintenanceUsecases.ErrExecutionNotFound))
			})
		})

		When("execution is already completed", func() {
			It("should return an error", func() {
				execution.MarkCompleted(completedBy)
				mockRepository.EXPECT().
					GetByID(gomock.Any(), execution.ID).
					Return(execution, nil)

				err := service.MarkExecutionCompleted(context.Background(), execution.ID, completedBy)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("already completed"))
			})
		})

		When("execution is deleted", func() {
			It("should return an error", func() {
				execution.SoftDelete()
				mockRepository.EXPECT().
					GetByID(gomock.Any(), execution.ID).
					Return(execution, nil)

				err := service.MarkExecutionCompleted(context.Background(), execution.ID, completedBy)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("deleted"))
			})
		})
	})
})
