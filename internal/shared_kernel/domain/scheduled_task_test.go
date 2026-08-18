package domain_test

import (
	"time"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("ScheduledTask", func() {
	ginkgo.Context("CalculateNextExecution", func() {
		ginkgo.When("using interval-based scheduling with daily execution", func() {
			var scheduledTask domain.ScheduledTask
			var initialDay time.Time
			var dayInterval int
			var executionTime string
			var timezone string

			ginkgo.BeforeEach(func() {
				// Setup: Daily task at 01:00 starting from 2025-10-11
				initialDay, _ = time.Parse(time.RFC3339, "2025-10-11T00:00:00Z")
				dayInterval = 1
				executionTime = "01:00"
				timezone = "America/Argentina/Buenos_Aires" // UTC-3

				tenant := domain.Tenant{
					ID:   domain.ID("test-tenant"),
					Name: "Test Tenant",
				}
				device := domain.Device{
					ID:   domain.ID("test-device"),
					Name: "Test Device",
				}

				// Construct scheduled task directly (bypassing builder validation for testing)
				scheduledTask = domain.ScheduledTask{
					ID:      domain.ID(utils.GenerateUUID()),
					Version: 1,
					Tenant:  tenant,
					Device:  device,
					CommandTemplates: []domain.CommandTemplate{
						{
							Device: device,
							Payload: domain.CommandPayload{
								Index: 1,
								Value: 10,
							},
							Priority: domain.CommandPriority("NORMAL"),
							WaitFor:  0,
						},
					},
					Scheduling: domain.SchedulingConfiguration{
						Type:          domain.SchedulingTypeInterval,
						InitialDay:    &utils.Time{Time: initialDay},
						DayInterval:   &dayInterval,
						ExecutionTime: &executionTime,
					},
					IsActive:  true,
					CreatedAt: utils.Time{Time: time.Now()},
					UpdatedAt: utils.Time{Time: time.Now()},
				}
			})

			ginkgo.It("should calculate next execution correctly when never executed", func() {
				// LastExecutedAt is nil, so should use InitialDay
				nextExecution, err := scheduledTask.CalculateNextExecution(timezone)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Expected: 2025-10-11 01:00:00 in Buenos Aires timezone
				location, _ := time.LoadLocation(timezone)
				expected := time.Date(2025, 10, 11, 1, 0, 0, 0, location)

				gomega.Expect(nextExecution.Equal(expected)).To(gomega.BeTrue(),
					"Expected %v but got %v", expected, nextExecution)
			})

			ginkgo.It("should calculate next execution correctly after being executed in the afternoon", func() {
				// This reproduces the production bug:
				// Simulate: Last executed on 2025-10-14 at 14:56:26 (Buenos Aires time UTC-3)
				location, _ := time.LoadLocation(timezone)
				lastExecuted := time.Date(2025, 10, 14, 14, 56, 26, 0, location)
				scheduledTask.LastExecutedAt = &utils.Time{Time: lastExecuted}

				nextExecution, err := scheduledTask.CalculateNextExecution(timezone)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Expected: 2025-10-15 01:00:00 in Buenos Aires timezone (next day at execution time)
				expected := time.Date(2025, 10, 15, 1, 0, 0, 0, location)

				// Verify the next execution is in the future (not in the past)
				now := time.Date(2025, 10, 14, 15, 0, 0, 0, location) // Current time is 15:00 same day
				gomega.Expect(nextExecution.After(now)).To(gomega.BeTrue(),
					"Next execution %v should be in the future, not in the past. Current time: %v",
					nextExecution, now)

				// Verify it's the correct date and time
				gomega.Expect(nextExecution.Equal(expected)).To(gomega.BeTrue(),
					"Expected %v but got %v", expected, nextExecution)
			})

			ginkgo.It("should not execute again if last execution was very recent (within same day)", func() {
				// Simulate: Last executed on 2025-10-14 at 01:05 (5 minutes after scheduled time)
				location, _ := time.LoadLocation(timezone)
				lastExecuted := time.Date(2025, 10, 14, 1, 5, 0, 0, location)
				scheduledTask.LastExecutedAt = &utils.Time{Time: lastExecuted}

				nextExecution, err := scheduledTask.CalculateNextExecution(timezone)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Expected: 2025-10-15 01:00:00 (next day, not same day)
				expected := time.Date(2025, 10, 15, 1, 0, 0, 0, location)

				gomega.Expect(nextExecution.Equal(expected)).To(gomega.BeTrue(),
					"Expected %v but got %v", expected, nextExecution)

				// Verify that if current time is still on 2025-10-14, shouldExecuteSchedule would return false
				currentTime := time.Date(2025, 10, 14, 2, 0, 0, 0, location)
				shouldExecute := nextExecution.Before(currentTime) || nextExecution.Equal(currentTime)
				gomega.Expect(shouldExecute).To(gomega.BeFalse(),
					"Should NOT execute again on the same day")
			})
		})
	})
})
