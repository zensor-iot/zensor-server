package domain_test

import (
	"fmt"
	"time"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("ScheduledTask", func() {
	ginkgo.Context("CalculateNextExecution", func() {
		var (
			scheduledTask domain.ScheduledTask
			initialDay    time.Time
		)

		ginkgo.BeforeEach(func() {
			initialDay = time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
			scheduledTask = domain.ScheduledTask{
				ID: domain.ID("test-scheduled-task"),
				Tenant: domain.Tenant{
					ID: domain.ID("test-tenant"),
				},
				Device: domain.Device{
					ID: domain.ID("test-device"),
				},
				CommandTemplates: []domain.CommandTemplate{},
				CreatedAt:        utils.Time{Time: initialDay},
			}
		})

		ginkgo.When("using interval scheduling", func() {
			ginkgo.BeforeEach(func() {
				executionTime := "02:00"
				dayInterval := 2
				scheduledTask.Scheduling = domain.SchedulingConfiguration{
					Type:          domain.SchedulingTypeInterval,
					InitialDay:    &utils.Time{Time: initialDay},
					DayInterval:   &dayInterval,
					ExecutionTime: &executionTime,
				}
			})

			ginkgo.It("should calculate next execution for first run", func() {
				nextExec, err := scheduledTask.CalculateNextExecution("UTC")

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				expected := time.Date(2024, 1, 15, 2, 0, 0, 0, time.UTC)
				gomega.Expect(nextExec).To(gomega.Equal(expected))
			})

			ginkgo.It("should calculate next execution after last execution", func() {
				lastExecuted := time.Date(2024, 1, 15, 2, 0, 0, 0, time.UTC)
				scheduledTask.LastExecutedAt = &utils.Time{Time: lastExecuted}

				nextExec, err := scheduledTask.CalculateNextExecution("UTC")

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				expected := time.Date(2024, 1, 17, 2, 0, 0, 0, time.UTC) // 2 days later
				gomega.Expect(nextExec).To(gomega.Equal(expected))
			})

			ginkgo.It("should calculate next execution for different timezone", func() {
				nextExec, err := scheduledTask.CalculateNextExecution("America/New_York")

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				// Should be 2:00 AM in the specified timezone
				gomega.Expect(nextExec.Hour()).To(gomega.Equal(2))
			})

			ginkgo.It("should handle 3-day intervals", func() {
				dayInterval := 3
				scheduledTask.Scheduling.DayInterval = &dayInterval
				lastExecuted := time.Date(2024, 1, 15, 2, 0, 0, 0, time.UTC)
				scheduledTask.LastExecutedAt = &utils.Time{Time: lastExecuted}

				nextExec, err := scheduledTask.CalculateNextExecution("UTC")

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				expected := time.Date(2024, 1, 18, 2, 0, 0, 0, time.UTC) // 3 days later
				gomega.Expect(nextExec).To(gomega.Equal(expected))
			})

			ginkgo.It("should handle different execution times", func() {
				executionTime := "14:30"
				scheduledTask.Scheduling.ExecutionTime = &executionTime
				lastExecuted := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
				scheduledTask.LastExecutedAt = &utils.Time{Time: lastExecuted}

				nextExec, err := scheduledTask.CalculateNextExecution("UTC")

				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				expected := time.Date(2024, 1, 17, 14, 30, 0, 0, time.UTC) // 2 days later at 2:30 PM
				gomega.Expect(nextExec).To(gomega.Equal(expected))
			})
		})

		ginkgo.When("using invalid scheduling type", func() {
			ginkgo.It("should return error for cron scheduling type", func() {
				scheduledTask.Scheduling = domain.SchedulingConfiguration{
					Type: domain.SchedulingTypeCron,
				}

				_, err := scheduledTask.CalculateNextExecution("UTC")

				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("calculateNextExecution only supports interval scheduling"))
			})
		})
	})

	ginkgo.Context("ScheduledTaskBuilder with interval scheduling", func() {
		ginkgo.When("creating scheduled task with interval scheduling", func() {
			ginkgo.It("should create scheduled task with interval scheduling", func() {
				// Given: A scheduled task with interval scheduling
				initialDay := time.Now().AddDate(0, 0, 1) // Tomorrow
				schedulingConfig := domain.SchedulingConfiguration{
					Type:          domain.SchedulingTypeInterval,
					InitialDay:    &utils.Time{Time: initialDay},
					DayInterval:   intPtr(2),
					ExecutionTime: stringPtr("02:00"),
				}

				// When: Creating scheduled task
				commandTemplate := domain.CommandTemplate{
					Device:   domain.Device{ID: domain.ID("test-device")},
					Port:     1,
					Priority: domain.CommandPriority("normal"),
					Payload: domain.CommandPayload{
						Index: domain.Index(1),
						Value: domain.CommandValue(100),
					},
					WaitFor: 0,
				}

				scheduledTask, err := domain.NewScheduledTaskBuilder().
					WithTenant(domain.Tenant{ID: domain.ID("test-tenant")}).
					WithDevice(domain.Device{ID: domain.ID("test-device")}).
					WithCommandTemplates([]domain.CommandTemplate{commandTemplate}).
					WithScheduling(schedulingConfig).
					WithIsActive(true).
					Build()

				// Then: Should create successfully
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(scheduledTask.Scheduling.Type).To(gomega.Equal(domain.SchedulingTypeInterval))
				gomega.Expect(scheduledTask.Scheduling.InitialDay).NotTo(gomega.BeNil())
				gomega.Expect(scheduledTask.Scheduling.InitialDay.Time).To(gomega.Equal(initialDay))
				gomega.Expect(scheduledTask.Scheduling.DayInterval).NotTo(gomega.BeNil())
				gomega.Expect(*scheduledTask.Scheduling.DayInterval).To(gomega.Equal(2))
				gomega.Expect(scheduledTask.Scheduling.ExecutionTime).NotTo(gomega.BeNil())
				gomega.Expect(*scheduledTask.Scheduling.ExecutionTime).To(gomega.Equal("02:00"))
			})
		})

		ginkgo.When("validating interval scheduling requirements", func() {
			ginkgo.It("should fail validation when initial_day is missing", func() {
				// Given: A scheduled task builder with interval scheduling but missing initial_day
				commandTemplate := domain.CommandTemplate{
					Device:   domain.Device{ID: domain.ID("test-device")},
					Port:     1,
					Priority: domain.CommandPriority("normal"),
					Payload: domain.CommandPayload{
						Index: domain.Index(1),
						Value: domain.CommandValue(100),
					},
					WaitFor: 0,
				}

				_, err := domain.NewScheduledTaskBuilder().
					WithTenant(domain.Tenant{ID: domain.ID("test-tenant")}).
					WithDevice(domain.Device{ID: domain.ID("test-device")}).
					WithCommandTemplates([]domain.CommandTemplate{commandTemplate}).
					WithScheduling(domain.SchedulingConfiguration{
						Type:          domain.SchedulingTypeInterval,
						DayInterval:   intPtr(2),
						ExecutionTime: stringPtr("02:00"),
					}).
					WithIsActive(true).
					Build()

				// Then: Should fail validation
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("initial_day is required for interval scheduling"))
			})

			ginkgo.It("should fail validation when day_interval is missing", func() {
				// Given: A scheduled task builder with interval scheduling but missing day_interval
				commandTemplate := domain.CommandTemplate{
					Device:   domain.Device{ID: domain.ID("test-device")},
					Port:     1,
					Priority: domain.CommandPriority("normal"),
					Payload: domain.CommandPayload{
						Index: domain.Index(1),
						Value: domain.CommandValue(100),
					},
					WaitFor: 0,
				}

				_, err := domain.NewScheduledTaskBuilder().
					WithTenant(domain.Tenant{ID: domain.ID("test-tenant")}).
					WithDevice(domain.Device{ID: domain.ID("test-device")}).
					WithCommandTemplates([]domain.CommandTemplate{commandTemplate}).
					WithScheduling(domain.SchedulingConfiguration{
						Type:          domain.SchedulingTypeInterval,
						InitialDay:    &utils.Time{Time: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
						ExecutionTime: stringPtr("02:00"),
					}).
					WithIsActive(true).
					Build()

				// Then: Should fail validation
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("day_interval must be greater than 0 for interval scheduling"))
			})

			ginkgo.It("should fail validation when execution_time is missing", func() {
				// Given: A scheduled task builder with interval scheduling but missing execution_time
				commandTemplate := domain.CommandTemplate{
					Device:   domain.Device{ID: domain.ID("test-device")},
					Port:     1,
					Priority: domain.CommandPriority("normal"),
					Payload: domain.CommandPayload{
						Index: domain.Index(1),
						Value: domain.CommandValue(100),
					},
					WaitFor: 0,
				}

				_, err := domain.NewScheduledTaskBuilder().
					WithTenant(domain.Tenant{ID: domain.ID("test-tenant")}).
					WithDevice(domain.Device{ID: domain.ID("test-device")}).
					WithCommandTemplates([]domain.CommandTemplate{commandTemplate}).
					WithScheduling(domain.SchedulingConfiguration{
						Type:        domain.SchedulingTypeInterval,
						InitialDay:  &utils.Time{Time: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
						DayInterval: intPtr(2),
					}).
					WithIsActive(true).
					Build()

				// Then: Should fail validation
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("execution_time is required for interval scheduling"))
			})
		})

		ginkgo.When("maintaining backward compatibility", func() {
			ginkgo.It("should create scheduled task with legacy cron format", func() {
				// Given: A scheduled task with legacy cron format
				commandTemplate := domain.CommandTemplate{
					Device:   domain.Device{ID: domain.ID("test-device")},
					Port:     1,
					Priority: domain.CommandPriority("normal"),
					Payload: domain.CommandPayload{
						Index: domain.Index(1),
						Value: domain.CommandValue(100),
					},
					WaitFor: 0,
				}

				scheduledTask, err := domain.NewScheduledTaskBuilder().
					WithTenant(domain.Tenant{ID: domain.ID("test-tenant")}).
					WithDevice(domain.Device{ID: domain.ID("test-device")}).
					WithCommandTemplates([]domain.CommandTemplate{commandTemplate}).
					WithSchedule("0 0 * * *").
					WithIsActive(true).
					Build()

				// Then: Should create successfully with cron scheduling type
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(scheduledTask.Schedule).To(gomega.Equal("0 0 * * *"))
				gomega.Expect(scheduledTask.Scheduling.Type).To(gomega.Equal(domain.SchedulingTypeCron))
			})
		})
	})

	ginkgo.Context("Interval calculation scenarios", func() {
		ginkgo.When("calculating next execution with different intervals", func() {
			ginkgo.It("should handle daily interval correctly", func() {
				// Given: A scheduled task with 1-day interval (daily)
				initialDay := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
				scheduledTask := domain.ScheduledTask{
					ID:               domain.ID("test-scheduled-task"),
					Tenant:           domain.Tenant{ID: domain.ID("test-tenant")},
					Device:           domain.Device{ID: domain.ID("test-device")},
					CommandTemplates: []domain.CommandTemplate{},
					Scheduling: domain.SchedulingConfiguration{
						Type:          domain.SchedulingTypeInterval,
						InitialDay:    &utils.Time{Time: initialDay},
						DayInterval:   intPtr(1),
						ExecutionTime: stringPtr("01:00"),
					},
					CreatedAt: utils.Time{Time: initialDay},
				}

				// When: Calculating next execution time
				nextExec, err := scheduledTask.CalculateNextExecution("UTC")

				// Then: Should calculate next day at 1:00 AM
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				expected := time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC)
				gomega.Expect(nextExec).To(gomega.Equal(expected))
			})

			ginkgo.It("should handle 3-day interval correctly", func() {
				// Given: A scheduled task with 3-day interval
				initialDay := time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC)
				scheduledTask := domain.ScheduledTask{
					ID:               domain.ID("test-scheduled-task"),
					Tenant:           domain.Tenant{ID: domain.ID("test-tenant")},
					Device:           domain.Device{ID: domain.ID("test-device")},
					CommandTemplates: []domain.CommandTemplate{},
					Scheduling: domain.SchedulingConfiguration{
						Type:          domain.SchedulingTypeInterval,
						InitialDay:    &utils.Time{Time: initialDay},
						DayInterval:   intPtr(3),
						ExecutionTime: stringPtr("15:00"),
					},
					CreatedAt: utils.Time{Time: initialDay},
				}

				// When: Setting last executed time and calculating next execution
				lastExecuted := time.Date(2024, 1, 10, 15, 0, 0, 0, time.UTC)
				scheduledTask.LastExecutedAt = &utils.Time{Time: lastExecuted}

				nextExec, err := scheduledTask.CalculateNextExecution("UTC")

				// Then: Should calculate next execution 3 days later at 3:00 PM
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				expected := time.Date(2024, 1, 13, 15, 0, 0, 0, time.UTC)
				gomega.Expect(nextExec).To(gomega.Equal(expected))
			})
		})

		ginkgo.When("evaluating multiple execution scenarios", func() {
			var (
				scheduledTask domain.ScheduledTask
				initialDay    time.Time
			)

			ginkgo.BeforeEach(func() {
				initialDay = time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
				scheduledTask = domain.ScheduledTask{
					ID:               domain.ID("test-scheduled-task"),
					Tenant:           domain.Tenant{ID: domain.ID("test-tenant")},
					Device:           domain.Device{ID: domain.ID("test-device")},
					CommandTemplates: []domain.CommandTemplate{},
					Scheduling: domain.SchedulingConfiguration{
						Type:          domain.SchedulingTypeInterval,
						InitialDay:    &utils.Time{Time: initialDay},
						DayInterval:   intPtr(2),
						ExecutionTime: stringPtr("02:00"),
					},
					CreatedAt: utils.Time{Time: initialDay},
				}
			})

			ginkgo.It("should calculate first execution correctly", func() {
				// When: Calculating first execution
				nextExec, err := scheduledTask.CalculateNextExecution("UTC")

				// Then: Should be first execution at initial day
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				expected := time.Date(2024, 1, 15, 2, 0, 0, 0, time.UTC)
				gomega.Expect(nextExec).To(gomega.Equal(expected))
			})

			ginkgo.It("should calculate execution after first run", func() {
				// Given: Last executed time is set
				lastExecuted := time.Date(2024, 1, 15, 2, 0, 0, 0, time.UTC)
				scheduledTask.LastExecutedAt = &utils.Time{Time: lastExecuted}

				// When: Calculating next execution
				nextExec, err := scheduledTask.CalculateNextExecution("UTC")

				// Then: Should be 2 days later
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				expected := time.Date(2024, 1, 17, 2, 0, 0, 0, time.UTC)
				gomega.Expect(nextExec).To(gomega.Equal(expected))
			})

			ginkgo.It("should calculate execution after second run", func() {
				// Given: Last executed time is set to second execution
				lastExecuted := time.Date(2024, 1, 17, 2, 0, 0, 0, time.UTC)
				scheduledTask.LastExecutedAt = &utils.Time{Time: lastExecuted}

				// When: Calculating next execution
				nextExec, err := scheduledTask.CalculateNextExecution("UTC")

				// Then: Should be 2 days later
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				expected := time.Date(2024, 1, 19, 2, 0, 0, 0, time.UTC)
				gomega.Expect(nextExec).To(gomega.Equal(expected))
			})
		})
	})

	ginkgo.Context("ScheduledTaskBuilder validation", func() {
		ginkgo.When("validating interval scheduling with past initial day", func() {
			ginkgo.It("should return error when initial day + execution time is in the past", func() {
				// Given: A past initial day and execution time
				pastDate := time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC)
				executionTime := "02:00"
				dayInterval := 2

				// When: Building a scheduled task with past initial day
				commandTemplate, _ := domain.NewCommandTemplateBuilder().
					WithDevice(domain.Device{ID: domain.ID("test-device")}).
					WithPayload(domain.CommandPayload{Index: domain.Index(1), Value: domain.CommandValue(100)}).
					WithPriority(domain.CommandPriority("normal")).
					WithWaitFor(0).
					Build()

				_, err := domain.NewScheduledTaskBuilder().
					WithTenant(domain.Tenant{ID: domain.ID("test-tenant")}).
					WithDevice(domain.Device{ID: domain.ID("test-device")}).
					WithCommandTemplates([]domain.CommandTemplate{commandTemplate}).
					WithScheduling(domain.SchedulingConfiguration{
						Type:          domain.SchedulingTypeInterval,
						InitialDay:    &utils.Time{Time: pastDate},
						DayInterval:   &dayInterval,
						ExecutionTime: &executionTime,
					}).
					Build()

				// Then: Should return validation error
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("initial_day with execution_time must be in the future"))
			})

			ginkgo.It("should allow future initial day + execution time", func() {
				// Given: A future initial day and execution time
				futureDate := time.Now().AddDate(0, 0, 1) // Tomorrow
				executionTime := "02:00"
				dayInterval := 2

				// When: Building a scheduled task with future initial day
				commandTemplate, _ := domain.NewCommandTemplateBuilder().
					WithDevice(domain.Device{ID: domain.ID("test-device")}).
					WithPayload(domain.CommandPayload{Index: domain.Index(1), Value: domain.CommandValue(100)}).
					WithPriority(domain.CommandPriority("normal")).
					WithWaitFor(0).
					Build()

				_, err := domain.NewScheduledTaskBuilder().
					WithTenant(domain.Tenant{ID: domain.ID("test-tenant")}).
					WithDevice(domain.Device{ID: domain.ID("test-device")}).
					WithCommandTemplates([]domain.CommandTemplate{commandTemplate}).
					WithScheduling(domain.SchedulingConfiguration{
						Type:          domain.SchedulingTypeInterval,
						InitialDay:    &utils.Time{Time: futureDate},
						DayInterval:   &dayInterval,
						ExecutionTime: &executionTime,
					}).
					Build()

				// Then: Should succeed
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should return error when initial day is today but execution time is in the past", func() {
				// Given: Today's date but with a past execution time
				today := time.Now().Truncate(24 * time.Hour)
				pastTime := time.Now().Add(-2 * time.Hour) // 2 hours ago
				executionTime := fmt.Sprintf("%02d:%02d", pastTime.Hour(), pastTime.Minute())
				dayInterval := 2

				// When: Building a scheduled task with past execution time today
				commandTemplate, _ := domain.NewCommandTemplateBuilder().
					WithDevice(domain.Device{ID: domain.ID("test-device")}).
					WithPayload(domain.CommandPayload{Index: domain.Index(1), Value: domain.CommandValue(100)}).
					WithPriority(domain.CommandPriority("normal")).
					WithWaitFor(0).
					Build()

				_, err := domain.NewScheduledTaskBuilder().
					WithTenant(domain.Tenant{ID: domain.ID("test-tenant")}).
					WithDevice(domain.Device{ID: domain.ID("test-device")}).
					WithCommandTemplates([]domain.CommandTemplate{commandTemplate}).
					WithScheduling(domain.SchedulingConfiguration{
						Type:          domain.SchedulingTypeInterval,
						InitialDay:    &utils.Time{Time: today},
						DayInterval:   &dayInterval,
						ExecutionTime: &executionTime,
					}).
					Build()

				// Then: Should return validation error
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("initial_day with execution_time must be in the future"))
			})
		})
	})
})

// Helper functions
func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
