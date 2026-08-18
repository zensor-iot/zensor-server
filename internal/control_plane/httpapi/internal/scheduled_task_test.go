package internal_test

import (
	"encoding/json"
	"time"
	"zensor-server/internal/control_plane/httpapi/internal"
	"zensor-server/internal/infra/utils"
	"zensor-server/internal/shared_kernel/domain"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SchedulingConfigurationRequest", func() {
	Context("ToSchedulingConfiguration", func() {
		var request internal.SchedulingConfigurationRequest
		var expected domain.SchedulingConfiguration
		var result domain.SchedulingConfiguration

		When("interval scheduling configuration", func() {
			BeforeEach(func() {
				request = internal.SchedulingConfigurationRequest{
					Type:          "interval",
					InitialDay:    timePtr(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)),
					DayInterval:   intPtr(2),
					ExecutionTime: stringPtr("02:00"),
				}
				expected = domain.SchedulingConfiguration{
					Type:          domain.SchedulingTypeInterval,
					InitialDay:    &utils.Time{Time: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
					DayInterval:   intPtr(2),
					ExecutionTime: stringPtr("02:00"),
				}
				result = request.ToSchedulingConfiguration()
			})

			It("should convert type correctly", func() {
				Expect(result.Type).To(Equal(expected.Type))
			})

			It("should convert initial day correctly", func() {
				Expect(result.InitialDay).NotTo(BeNil())
				Expect(result.InitialDay.Time).To(Equal(expected.InitialDay.Time))
			})

			It("should convert day interval correctly", func() {
				Expect(result.DayInterval).To(Equal(expected.DayInterval))
			})

			It("should convert execution time correctly", func() {
				Expect(result.ExecutionTime).To(Equal(expected.ExecutionTime))
			})
		})

		When("cron scheduling configuration", func() {
			BeforeEach(func() {
				request = internal.SchedulingConfigurationRequest{
					Type:     "cron",
					Schedule: stringPtr("0 0 * * *"),
				}
				expected = domain.SchedulingConfiguration{
					Type: domain.SchedulingTypeCron,
				}
				result = request.ToSchedulingConfiguration()
			})

			It("should convert type correctly", func() {
				Expect(result.Type).To(Equal(expected.Type))
			})

			It("should have nil initial day", func() {
				Expect(result.InitialDay).To(BeNil())
			})

			It("should have nil day interval", func() {
				Expect(result.DayInterval).To(BeNil())
			})

			It("should have nil execution time", func() {
				Expect(result.ExecutionTime).To(BeNil())
			})
		})

		When("empty configuration", func() {
			BeforeEach(func() {
				request = internal.SchedulingConfigurationRequest{
					Type: "",
				}
				expected = domain.SchedulingConfiguration{
					Type: "",
				}
				result = request.ToSchedulingConfiguration()
			})

			It("should have empty type", func() {
				Expect(result.Type).To(Equal(expected.Type))
			})

			It("should have nil initial day", func() {
				Expect(result.InitialDay).To(BeNil())
			})

			It("should have nil day interval", func() {
				Expect(result.DayInterval).To(BeNil())
			})

			It("should have nil execution time", func() {
				Expect(result.ExecutionTime).To(BeNil())
			})
		})
	})
})

var _ = Describe("FromSchedulingConfiguration", func() {
	var config domain.SchedulingConfiguration
	var nextExecution *time.Time
	var result *internal.SchedulingConfigurationResponse

	When("interval scheduling with next execution", func() {
		var initialDay time.Time
		var nextExecutionTime time.Time

		BeforeEach(func() {
			initialDay = time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
			nextExecutionTime = time.Date(2024, 1, 17, 2, 0, 0, 0, time.UTC)
			nextExecution = &nextExecutionTime

			config = domain.SchedulingConfiguration{
				Type:          domain.SchedulingTypeInterval,
				InitialDay:    &utils.Time{Time: initialDay},
				DayInterval:   intPtr(2),
				ExecutionTime: stringPtr("02:00"),
			}
			result = internal.FromSchedulingConfiguration(config, nextExecution)
		})

		It("should convert type correctly", func() {
			Expect(result.Type).To(Equal("interval"))
		})

		It("should convert initial day correctly", func() {
			Expect(result.InitialDay).NotTo(BeNil())
			Expect(*result.InitialDay).To(Equal(initialDay))
		})

		It("should convert day interval correctly", func() {
			Expect(result.DayInterval).NotTo(BeNil())
			Expect(*result.DayInterval).To(Equal(2))
		})

		It("should convert execution time correctly", func() {
			Expect(result.ExecutionTime).NotTo(BeNil())
			Expect(*result.ExecutionTime).To(Equal("02:00"))
		})

		It("should convert next execution correctly", func() {
			Expect(result.NextExecution).NotTo(BeNil())
			Expect(*result.NextExecution).To(Equal(nextExecutionTime))
		})
	})

	When("cron scheduling without next execution", func() {
		BeforeEach(func() {
			nextExecution = nil
			config = domain.SchedulingConfiguration{
				Type: domain.SchedulingTypeCron,
			}
			result = internal.FromSchedulingConfiguration(config, nextExecution)
		})

		It("should convert type correctly", func() {
			Expect(result.Type).To(Equal("cron"))
		})

		It("should have nil initial day", func() {
			Expect(result.InitialDay).To(BeNil())
		})

		It("should have nil day interval", func() {
			Expect(result.DayInterval).To(BeNil())
		})

		It("should have nil execution time", func() {
			Expect(result.ExecutionTime).To(BeNil())
		})

		It("should have nil next execution", func() {
			Expect(result.NextExecution).To(BeNil())
		})
	})
})

var _ = Describe("ScheduledTaskCreateRequest", func() {
	Context("JSONSerialization", func() {
		var request internal.ScheduledTaskCreateRequest
		var initialDay time.Time

		When("marshaling and unmarshaling", func() {
			BeforeEach(func() {
				initialDay = time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

				request = internal.ScheduledTaskCreateRequest{
					Commands: []internal.CommandSendPayloadRequest{
						{
							Index:    1,
							Value:    100,
							Priority: "NORMAL",
							WaitFor:  utils.Duration(0),
						},
					},
					Scheduling: &internal.SchedulingConfigurationRequest{
						Type:          "interval",
						InitialDay:    &initialDay,
						DayInterval:   intPtr(2),
						ExecutionTime: stringPtr("02:00"),
					},
					IsActive: true,
				}
			})

			It("should marshal and unmarshal correctly", func() {
				jsonData, err := json.Marshal(request)
				Expect(err).NotTo(HaveOccurred())

				var unmarshaled internal.ScheduledTaskCreateRequest
				err = json.Unmarshal(jsonData, &unmarshaled)
				Expect(err).NotTo(HaveOccurred())

				Expect(unmarshaled.Commands).To(Equal(request.Commands))
				Expect(unmarshaled.IsActive).To(Equal(request.IsActive))
				Expect(unmarshaled.Scheduling).NotTo(BeNil())
				Expect(unmarshaled.Scheduling.Type).To(Equal(request.Scheduling.Type))
				Expect(unmarshaled.Scheduling.InitialDay).To(Equal(request.Scheduling.InitialDay))
				Expect(unmarshaled.Scheduling.DayInterval).To(Equal(request.Scheduling.DayInterval))
				Expect(unmarshaled.Scheduling.ExecutionTime).To(Equal(request.Scheduling.ExecutionTime))
			})
		})
	})
})

var _ = Describe("ScheduledTaskResponse", func() {
	Context("JSONSerialization", func() {
		var response internal.ScheduledTaskResponse
		var initialDay time.Time
		var nextExecution time.Time

		When("marshaling and unmarshaling", func() {
			BeforeEach(func() {
				initialDay = time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
				nextExecution = time.Date(2024, 1, 17, 2, 0, 0, 0, time.UTC)

				response = internal.ScheduledTaskResponse{
					ID:       "test-id",
					DeviceID: "test-device-id",
					Commands: []internal.CommandSendPayloadRequest{
						{
							Index:    1,
							Value:    100,
							Priority: "NORMAL",
							WaitFor:  utils.Duration(0),
						},
					},
					Scheduling: &internal.SchedulingConfigurationResponse{
						Type:          "interval",
						InitialDay:    &initialDay,
						DayInterval:   intPtr(2),
						ExecutionTime: stringPtr("02:00"),
						NextExecution: &nextExecution,
					},
					IsActive: true,
				}
			})

			It("should marshal and unmarshal correctly", func() {
				jsonData, err := json.Marshal(response)
				Expect(err).NotTo(HaveOccurred())

				var unmarshaled internal.ScheduledTaskResponse
				err = json.Unmarshal(jsonData, &unmarshaled)
				Expect(err).NotTo(HaveOccurred())

				Expect(unmarshaled.ID).To(Equal(response.ID))
				Expect(unmarshaled.DeviceID).To(Equal(response.DeviceID))
				Expect(unmarshaled.Commands).To(Equal(response.Commands))
				Expect(unmarshaled.IsActive).To(Equal(response.IsActive))
				Expect(unmarshaled.Scheduling).NotTo(BeNil())
				Expect(unmarshaled.Scheduling.Type).To(Equal(response.Scheduling.Type))
				Expect(unmarshaled.Scheduling.InitialDay).To(Equal(response.Scheduling.InitialDay))
				Expect(unmarshaled.Scheduling.DayInterval).To(Equal(response.Scheduling.DayInterval))
				Expect(unmarshaled.Scheduling.ExecutionTime).To(Equal(response.Scheduling.ExecutionTime))
				Expect(unmarshaled.Scheduling.NextExecution).To(Equal(response.Scheduling.NextExecution))
			})
		})
	})
})

func timePtr(t time.Time) *time.Time {
	return &t
}

func intPtr(i int) *int {
	return &i
}

func stringPtr(s string) *string {
	return &s
}
