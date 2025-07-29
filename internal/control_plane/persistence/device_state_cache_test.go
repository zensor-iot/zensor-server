package persistence_test

import (
	"context"
	"fmt"
	"zensor-server/internal/control_plane/persistence"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/data_plane/dto"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("SimpleDeviceStateCacheService", func() {
	var (
		cache usecases.DeviceStateCacheService
		ctx   context.Context
	)

	ginkgo.BeforeEach(func() {
		cache = persistence.NewSimpleDeviceStateCacheService()
		ctx = context.Background()
	})

	ginkgo.Context("NewSimpleDeviceStateCacheService", func() {
		ginkgo.When("creating a new device state cache service", func() {
			ginkgo.It("should create a valid cache instance", func() {
				gomega.Expect(cache).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Context("SetState", func() {
		var (
			deviceID   string
			sensorData map[string][]dto.SensorData
		)

		ginkgo.When("setting state for a device", func() {
			ginkgo.BeforeEach(func() {
				deviceID = "test-device-1"
				sensorData = map[string][]dto.SensorData{
					"temperature": {
						{Index: 0, Value: 25.5},
						{Index: 1, Value: 26.2},
					},
					"humidity": {
						{Index: 0, Value: 60.0},
					},
				}
			})

			ginkgo.It("should store the state successfully", func() {
				err := cache.SetState(ctx, deviceID, sensorData)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Verify the state was cached
				state, exists := cache.GetState(ctx, deviceID)
				gomega.Expect(exists).To(gomega.BeTrue())
				gomega.Expect(state.DeviceID).To(gomega.Equal(deviceID))
				gomega.Expect(state.Data).To(gomega.HaveLen(2))
				gomega.Expect(state.Data["temperature"]).To(gomega.HaveLen(2))
				gomega.Expect(state.Data["humidity"]).To(gomega.HaveLen(1))
				gomega.Expect(state.Data["temperature"][0].Value).To(gomega.Equal(25.5))
				gomega.Expect(state.Data["temperature"][1].Value).To(gomega.Equal(26.2))
				gomega.Expect(state.Data["humidity"][0].Value).To(gomega.Equal(60.0))
			})
		})

		ginkgo.When("overwriting existing state", func() {
			var (
				initialData map[string][]dto.SensorData
				updatedData map[string][]dto.SensorData
			)

			ginkgo.BeforeEach(func() {
				deviceID = "test-device"
				initialData = map[string][]dto.SensorData{
					"temperature": {{Index: 0, Value: 25.0}},
				}
				updatedData = map[string][]dto.SensorData{
					"temperature": {{Index: 0, Value: 30.0}},
					"humidity":    {{Index: 0, Value: 65.0}},
				}

				// Add initial state
				err := cache.SetState(ctx, deviceID, initialData)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should overwrite the existing state", func() {
				// Update with new data
				err := cache.SetState(ctx, deviceID, updatedData)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Verify the state was updated
				state, exists := cache.GetState(ctx, deviceID)
				gomega.Expect(exists).To(gomega.BeTrue())
				gomega.Expect(state.Data).To(gomega.HaveLen(2))
				gomega.Expect(state.Data["temperature"][0].Value).To(gomega.Equal(30.0))
				gomega.Expect(state.Data["humidity"][0].Value).To(gomega.Equal(65.0))
			})
		})
	})

	ginkgo.Context("GetState", func() {
		var deviceID string

		ginkgo.When("getting state for a non-existent device", func() {
			ginkgo.BeforeEach(func() {
				deviceID = "non-existent-device"
			})

			ginkgo.It("should return false and empty state", func() {
				state, exists := cache.GetState(ctx, deviceID)
				gomega.Expect(exists).To(gomega.BeFalse())
				gomega.Expect(state.DeviceID).To(gomega.BeEmpty())
			})
		})
	})

	ginkgo.Context("GetAllDeviceIDs", func() {
		var (
			device1Data map[string][]dto.SensorData
			device2Data map[string][]dto.SensorData
		)

		ginkgo.When("getting all device IDs", func() {
			ginkgo.BeforeEach(func() {
				device1Data = map[string][]dto.SensorData{
					"temperature": {{Index: 0, Value: 25.0}},
				}
				device2Data = map[string][]dto.SensorData{
					"humidity": {{Index: 0, Value: 70.0}},
				}

				// Add multiple device states
				err := cache.SetState(ctx, "device-1", device1Data)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				err = cache.SetState(ctx, "device-2", device2Data)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.It("should return all device IDs", func() {
				deviceIDs := cache.GetAllDeviceIDs(ctx)
				gomega.Expect(deviceIDs).To(gomega.HaveLen(2))
				gomega.Expect(deviceIDs).To(gomega.ContainElement("device-1"))
				gomega.Expect(deviceIDs).To(gomega.ContainElement("device-2"))
			})
		})
	})

	ginkgo.Context("ConcurrentAccess", func() {
		ginkgo.When("accessing the cache concurrently", func() {
			ginkgo.It("should handle concurrent updates safely", func() {
				// Test concurrent updates
				done := make(chan bool, 10)
				for i := 0; i < 10; i++ {
					go func(id int) {
						deviceID := fmt.Sprintf("device-%d", id)
						sensorData := map[string][]dto.SensorData{
							"temperature": {{Index: 0, Value: float64(id)}},
						}
						cache.SetState(ctx, deviceID, sensorData)
						done <- true
					}(i)
				}

				// Wait for all goroutines to complete
				for i := 0; i < 10; i++ {
					<-done
				}

				// Verify all states were cached
				deviceIDs := cache.GetAllDeviceIDs(ctx)
				gomega.Expect(deviceIDs).To(gomega.HaveLen(10))
			})
		})
	})
})
