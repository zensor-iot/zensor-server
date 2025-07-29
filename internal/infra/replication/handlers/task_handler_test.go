package handlers_test

import (
	"context"
	"errors"
	"zensor-server/internal/infra/replication/handlers"
	"zensor-server/internal/infra/sql"
	mocksql "zensor-server/test/unit/doubles/infra/sql"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("TaskHandler", func() {
	ginkgo.Context("Create", func() {
		var (
			ctrl    *gomock.Controller
			mockOrm *mocksql.MockORM
			handler *handlers.TaskHandler
			task    struct {
				ID       string
				DeviceID string
			}
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
			handler = handlers.NewTaskHandler(mockOrm)
			task = struct {
				ID       string
				DeviceID string
			}{ID: "task-1", DeviceID: "device-1"}
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.When("creating task successfully", func() {
			ginkgo.It("should create task without error", func() {
				// Set up mock expectations
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Create(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(nil)

				// Execute the method
				err := handler.Create(context.Background(), "task-1", task)

				// Assertions
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.When("creating task fails", func() {
			ginkgo.It("should return error when database fails", func() {
				// Set up mock expectations
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Create(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(errors.New("db error"))

				// Execute the method
				err := handler.Create(context.Background(), "task-1", task)

				// Assertions
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("creating task"))
			})
		})
	})

	ginkgo.Context("GetByID", func() {
		var (
			ctrl    *gomock.Controller
			mockOrm *mocksql.MockORM
			handler *handlers.TaskHandler
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
			handler = handlers.NewTaskHandler(mockOrm)
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.When("getting task successfully", func() {
			ginkgo.It("should return task data", func() {
				// Set up mock expectations
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().First(gomock.Any(), gomock.Any()).DoAndReturn(
					func(dest interface{}, conds ...interface{}) *mocksql.MockORM {
						// Set the destination with test data
						taskData := dest.(*handlers.TaskData)
						*taskData = handlers.TaskData{
							ID:       "task-1",
							DeviceID: "device-1",
							Version:  1,
						}
						return mockOrm
					},
				)
				mockOrm.EXPECT().Error().Return(nil)

				// Execute the method
				result, err := handler.GetByID(context.Background(), "task-1")

				// Assertions
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())

				resultMap, ok := result.(map[string]interface{})
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(resultMap["id"]).To(gomega.Equal("task-1"))
				gomega.Expect(resultMap["device_id"]).To(gomega.Equal("device-1"))
			})
		})

		ginkgo.When("task not found", func() {
			ginkgo.It("should return error when task not found", func() {
				// Set up mock expectations
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(sql.ErrRecordNotFound)

				// Execute the method
				result, err := handler.GetByID(context.Background(), "task-1")

				// Assertions
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("getting task"))
			})
		})
	})

	ginkgo.Context("Update", func() {
		var (
			ctrl    *gomock.Controller
			mockOrm *mocksql.MockORM
			handler *handlers.TaskHandler
			task    struct {
				ID       string
				DeviceID string
			}
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
			handler = handlers.NewTaskHandler(mockOrm)
			task = struct {
				ID       string
				DeviceID string
			}{ID: "task-1", DeviceID: "device-1"}
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.When("updating task successfully", func() {
			ginkgo.It("should update task without error", func() {
				// Set up mock expectations
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Save(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(nil)

				// Execute the method
				err := handler.Update(context.Background(), "task-1", task)

				// Assertions
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.When("updating task fails", func() {
			ginkgo.It("should return error when database fails", func() {
				// Set up mock expectations
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Save(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(errors.New("db error"))

				// Execute the method
				err := handler.Update(context.Background(), "task-1", task)

				// Assertions
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("updating task"))
			})
		})
	})
})
