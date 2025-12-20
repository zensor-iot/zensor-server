package usecases_test

import (
	"context"
	"errors"
	"time"
	"zensor-server/internal/infra/utils"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	maintenanceUsecases "zensor-server/internal/maintenance/usecases"
	shareddomain "zensor-server/internal/shared_kernel/domain"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MaintenanceExecutionService", func() {
	var service maintenanceUsecases.ExecutionService
	var mockRepository *mockMaintenanceExecutionRepository
	var mockActivityRepository *mockMaintenanceActivityRepository

	BeforeEach(func() {
		mockRepository = newMockMaintenanceExecutionRepository()
		mockActivityRepository = newMockMaintenanceActivityRepository()
		service = maintenanceUsecases.NewExecutionService(mockRepository, mockActivityRepository)
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

			mockActivityRepository.activities[activityID.String()] = activity

			execution, _ = maintenanceDomain.NewExecutionBuilder().
				WithActivityID(activityID).
				WithScheduledDate(time.Now().AddDate(0, 0, 30)).
				WithFieldValues(map[string]any{"maintenance_type": "Filter Replacement"}).
				Build()
		})

		When("creating a valid execution", func() {
			It("should successfully create the execution", func() {
				err := service.CreateExecution(context.Background(), execution)
				Expect(err).NotTo(HaveOccurred())
				Expect(mockRepository.createCalled).To(BeTrue())
				Expect(mockRepository.executions[execution.ID.String()]).To(Equal(execution))
			})
		})

		When("activity does not exist", func() {
			BeforeEach(func() {
				mockActivityRepository.getByIDError = maintenanceUsecases.ErrActivityNotFound
			})

			It("should return an error", func() {
				err := service.CreateExecution(context.Background(), execution)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("maintenance activity not found"))
			})
		})

		When("repository returns an error", func() {
			BeforeEach(func() {
				mockRepository.createError = errors.New("database error")
			})

			It("should return the error", func() {
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

			mockRepository.executions[executionID.String()] = execution
		})

		When("execution exists", func() {
			It("should return the execution", func() {
				result, err := service.GetExecution(context.Background(), executionID)
				Expect(err).NotTo(HaveOccurred())
				Expect(result.ID).To(Equal(executionID))
				Expect(result.ActivityID).To(Equal(execution.ActivityID))
			})
		})

		When("execution does not exist", func() {
			BeforeEach(func() {
				mockRepository.getByIDError = maintenanceUsecases.ErrExecutionNotFound
			})

			It("should return ErrMaintenanceExecutionNotFound", func() {
				result, err := service.GetExecution(context.Background(), executionID)
				Expect(err).To(MatchError(maintenanceUsecases.ErrExecutionNotFound))
				Expect(result.ID).To(BeEmpty())
			})
		})

		When("repository returns an error", func() {
			BeforeEach(func() {
				mockRepository.getByIDError = errors.New("database error")
			})

			It("should return the error", func() {
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

			mockRepository.executionsByActivity[activityID.String()] = executions
			mockRepository.totalByActivity[activityID.String()] = len(executions)
		})

		When("listing executions", func() {
			It("should return all executions for the activity", func() {
				pagination := maintenanceUsecases.Pagination{Limit: 10, Offset: 0}
				result, total, err := service.ListExecutionsByActivity(context.Background(), activityID, pagination)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(HaveLen(3))
				Expect(total).To(Equal(3))
			})
		})

		When("repository returns an error", func() {
			BeforeEach(func() {
				mockRepository.findAllByActivityError = errors.New("database error")
			})

			It("should return the error", func() {
				pagination := maintenanceUsecases.Pagination{Limit: 10, Offset: 0}
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

			mockRepository.executions[executionID.String()] = execution
		})

		When("marking completion for an existing execution", func() {
			It("should successfully mark the execution as completed", func() {
				err := service.MarkExecutionCompleted(context.Background(), execution.ID, completedBy)
				Expect(err).NotTo(HaveOccurred())
				Expect(mockRepository.markCompletedCalled).To(BeTrue())
			})
		})

		When("execution does not exist", func() {
			BeforeEach(func() {
				mockRepository.getByIDError = maintenanceUsecases.ErrExecutionNotFound
			})

			It("should return ErrMaintenanceExecutionNotFound", func() {
				err := service.MarkExecutionCompleted(context.Background(), execution.ID, completedBy)
				Expect(err).To(MatchError(maintenanceUsecases.ErrExecutionNotFound))
			})
		})

		When("execution is already completed", func() {
			BeforeEach(func() {
				execution.MarkCompleted(completedBy)
				mockRepository.executions[execution.ID.String()] = execution
			})

			It("should return an error", func() {
				err := service.MarkExecutionCompleted(context.Background(), execution.ID, completedBy)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("already completed"))
			})
		})

		When("execution is deleted", func() {
			BeforeEach(func() {
				execution.SoftDelete()
				mockRepository.executions[execution.ID.String()] = execution
			})

			It("should return an error", func() {
				err := service.MarkExecutionCompleted(context.Background(), execution.ID, completedBy)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("deleted"))
			})
		})
	})
})

type mockMaintenanceExecutionRepository struct {
	executions              map[string]maintenanceDomain.Execution
	executionsByActivity    map[string][]maintenanceDomain.Execution
	totalByActivity         map[string]int
	createCalled            bool
	getByIDCalled           bool
	findAllByActivityCalled bool
	updateCalled            bool
	markCompletedCalled     bool
	createError             error
	getByIDError            error
	findAllByActivityError  error
	updateError             error
	markCompletedError      error
}

func newMockMaintenanceExecutionRepository() *mockMaintenanceExecutionRepository {
	return &mockMaintenanceExecutionRepository{
		executions:           make(map[string]maintenanceDomain.Execution),
		executionsByActivity: make(map[string][]maintenanceDomain.Execution),
		totalByActivity:      make(map[string]int),
	}
}

func (m *mockMaintenanceExecutionRepository) Create(ctx context.Context, execution maintenanceDomain.Execution) error {
	m.createCalled = true
	if m.createError != nil {
		return m.createError
	}
	m.executions[execution.ID.String()] = execution
	return nil
}

func (m *mockMaintenanceExecutionRepository) GetByID(ctx context.Context, id shareddomain.ID) (maintenanceDomain.Execution, error) {
	m.getByIDCalled = true
	if m.getByIDError != nil {
		return maintenanceDomain.Execution{}, m.getByIDError
	}
	if execution, ok := m.executions[id.String()]; ok {
		return execution, nil
	}
	return maintenanceDomain.Execution{}, maintenanceUsecases.ErrExecutionNotFound
}

func (m *mockMaintenanceExecutionRepository) FindAllByActivity(ctx context.Context, activityID shareddomain.ID, pagination maintenanceUsecases.Pagination) ([]maintenanceDomain.Execution, int, error) {
	m.findAllByActivityCalled = true
	if m.findAllByActivityError != nil {
		return nil, 0, m.findAllByActivityError
	}
	if executions, ok := m.executionsByActivity[activityID.String()]; ok {
		total := m.totalByActivity[activityID.String()]
		return executions, total, nil
	}
	return []maintenanceDomain.Execution{}, 0, nil
}

func (m *mockMaintenanceExecutionRepository) Update(ctx context.Context, execution maintenanceDomain.Execution) error {
	m.updateCalled = true
	if m.updateError != nil {
		return m.updateError
	}
	m.executions[execution.ID.String()] = execution
	return nil
}

func (m *mockMaintenanceExecutionRepository) MarkCompleted(ctx context.Context, id shareddomain.ID, completedBy string) error {
	m.markCompletedCalled = true
	if m.markCompletedError != nil {
		return m.markCompletedError
	}
	execution := m.executions[id.String()]
	execution.MarkCompleted(completedBy)
	m.executions[id.String()] = execution
	return nil
}

func (m *mockMaintenanceExecutionRepository) FindAllOverdue(ctx context.Context, tenantID shareddomain.ID) ([]maintenanceDomain.Execution, error) {
	return nil, nil
}

func (m *mockMaintenanceExecutionRepository) FindAllDueSoon(ctx context.Context, tenantID shareddomain.ID, days int) ([]maintenanceDomain.Execution, error) {
	return nil, nil
}
