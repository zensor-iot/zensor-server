package handlers_test

import (
	"context"
	"errors"
	"time"
	"zensor-server/internal/infra/pubsub"
	"zensor-server/internal/infra/replication/handlers"
	"zensor-server/internal/infra/sql"
	mocksql "zensor-server/test/unit/doubles/infra/sql"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("DeviceHandler", func() {
	ginkgo.Context("NewDeviceHandler", func() {
		var (
			ctrl    *gomock.Controller
			mockOrm *mocksql.MockORM
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.It("should create a new device handler", func() {
			handler := handlers.NewDeviceHandler(mockOrm)
			gomega.Expect(handler).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Context("TopicName", func() {
		var (
			handler *handlers.DeviceHandler
		)

		ginkgo.BeforeEach(func() {
			handler = handlers.NewDeviceHandler(nil)
		})

		ginkgo.It("should return devices topic", func() {
			topic := handler.TopicName()
			gomega.Expect(topic).To(gomega.Equal(pubsub.Topic("devices")))
		})
	})

	ginkgo.Context("Create", func() {
		var (
			ctrl       *gomock.Controller
			mockOrm    *mocksql.MockORM
			handler    *handlers.DeviceHandler
			testDevice struct {
				ID          string
				Name        string
				DisplayName string
				AppEUI      string
				DevEUI      string
				AppKey      string
				TenantID    *string
				CreatedAt   time.Time
			}
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
			handler = handlers.NewDeviceHandler(mockOrm)
			testDevice = struct {
				ID          string
				Name        string
				DisplayName string
				AppEUI      string
				DevEUI      string
				AppKey      string
				TenantID    *string
				CreatedAt   time.Time
			}{
				ID:          "test-device-1",
				Name:        "Test Device",
				DisplayName: "Test Device Display",
				AppEUI:      "test-app-eui",
				DevEUI:      "test-dev-eui",
				AppKey:      "test-app-key",
				TenantID:    nil,
				CreatedAt:   time.Now(),
			}
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.When("creating device successfully", func() {
			ginkgo.It("should create device without error", func() {
				// Set up mock expectations
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Create(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(nil)

				// Execute the method
				err := handler.Create(context.Background(), "test-device-1", testDevice)

				// Assertions
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.When("creating device fails", func() {
			ginkgo.It("should return error when database fails", func() {
				// Set up mock expectations
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Create(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(errors.New("db error"))

				// Execute the method
				err := handler.Create(context.Background(), "test-device-1", testDevice)

				// Assertions
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("creating device"))
			})
		})
	})

	ginkgo.Context("GetByID", func() {
		var (
			ctrl    *gomock.Controller
			mockOrm *mocksql.MockORM
			handler *handlers.DeviceHandler
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
			handler = handlers.NewDeviceHandler(mockOrm)
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.When("getting device successfully", func() {
			ginkgo.It("should return device data", func() {
				// Set up mock expectations
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().First(gomock.Any(), gomock.Any()).DoAndReturn(
					func(dest interface{}, conds ...interface{}) *mocksql.MockORM {
						// Set the destination with test data
						deviceData := dest.(*handlers.DeviceData)
						*deviceData = handlers.DeviceData{
							ID:          "test-device-1",
							Name:        "Test Device",
							DisplayName: "Test Device Display",
							AppEUI:      "test-app-eui",
							DevEUI:      "test-dev-eui",
							AppKey:      "test-app-key",
							TenantID:    nil,
							CreatedAt:   time.Now(),
						}
						return mockOrm
					},
				)
				mockOrm.EXPECT().Error().Return(nil)

				// Execute the method
				result, err := handler.GetByID(context.Background(), "test-device-1")

				// Assertions
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())

				resultMap, ok := result.(map[string]interface{})
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(resultMap["id"]).To(gomega.Equal("test-device-1"))
				gomega.Expect(resultMap["name"]).To(gomega.Equal("Test Device"))
			})
		})

		ginkgo.When("device not found", func() {
			ginkgo.It("should return error when device not found", func() {
				// Set up mock expectations
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(sql.ErrRecordNotFound)

				// Execute the method
				result, err := handler.GetByID(context.Background(), "test-device-1")

				// Assertions
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("getting device"))
			})
		})
	})

	ginkgo.Context("Update", func() {
		var (
			ctrl       *gomock.Controller
			mockOrm    *mocksql.MockORM
			handler    *handlers.DeviceHandler
			testDevice struct {
				ID          string
				Name        string
				DisplayName string
			}
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
			handler = handlers.NewDeviceHandler(mockOrm)
			testDevice = struct {
				ID          string
				Name        string
				DisplayName string
			}{
				ID:          "test-device-1",
				Name:        "Updated Device",
				DisplayName: "Updated Device Display",
			}
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.When("updating device successfully", func() {
			ginkgo.It("should update device without error", func() {
				// Set up mock expectations - WithContext is called once in Update method
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Save(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(nil)

				// Execute the method
				err := handler.Update(context.Background(), "test-device-1", testDevice)

				// Assertions
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.When("updating device fails", func() {
			ginkgo.It("should return error when database fails", func() {
				// Set up mock expectations - WithContext is called once before Save
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Save(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(errors.New("db error"))

				// Execute the method
				err := handler.Update(context.Background(), "test-device-1", testDevice)

				// Assertions
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("updating device"))
			})
		})
	})
})
