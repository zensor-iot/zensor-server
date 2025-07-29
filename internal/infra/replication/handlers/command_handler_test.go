package handlers

import (
	"context"
	"errors"

	"zensor-server/internal/infra/sql"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = ginkgo.Describe("CommandHandler", func() {
	ginkgo.Context("Create", func() {
		var (
			orm     *MockORM
			handler *CommandHandler
			mockOrm *MockORM
			cmd     struct {
				ID         string
				DeviceName string
				DeviceID   string
				TaskID     string
			}
		)

		ginkgo.BeforeEach(func() {
			orm = &MockORM{}
			handler = NewCommandHandler(orm)
			mockOrm = &MockORM{}
			cmd = struct {
				ID         string
				DeviceName string
				DeviceID   string
				TaskID     string
			}{ID: "cmd-1", DeviceName: "dev", DeviceID: "dev-1", TaskID: "task-1"}
		})

		ginkgo.When("creating command successfully", func() {
			ginkgo.It("should create command without error", func() {
				mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
				mockOrm.On("Create", mock.Anything).Return(mockOrm)
				mockOrm.On("Error").Return(nil)
				handler.orm = mockOrm

				err := handler.Create(context.Background(), "cmd-1", cmd)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				mockOrm.AssertExpectations(ginkgo.GinkgoT())
			})
		})

		ginkgo.When("creating command with database error", func() {
			ginkgo.It("should return error", func() {
				mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
				mockOrm.On("Create", mock.Anything).Return(mockOrm)
				mockOrm.On("Error").Return(errors.New("db error"))
				handler.orm = mockOrm

				err := handler.Create(context.Background(), "cmd-1", cmd)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("creating command"))
				mockOrm.AssertExpectations(ginkgo.GinkgoT())
			})
		})
	})

	ginkgo.Context("GetByID", func() {
		var (
			orm     *MockORM
			handler *CommandHandler
			mockOrm *MockORM
		)

		ginkgo.BeforeEach(func() {
			orm = &MockORM{}
			handler = NewCommandHandler(orm)
			mockOrm = &MockORM{}
		})

		ginkgo.When("getting command by ID successfully", func() {
			ginkgo.It("should return command data", func() {
				mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
				mockOrm.On("First", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					dest := args.Get(0).(*CommandData)
					*dest = CommandData{
						ID:         "cmd-1",
						DeviceName: "dev",
						DeviceID:   "device-1",
						TaskID:     "task-1",
						Payload:    CommandPayload{Index: 1, Data: 100},
					}
				}).Return(mockOrm)
				mockOrm.On("Error").Return(nil)
				handler.orm = mockOrm

				result, err := handler.GetByID(context.Background(), "cmd-1")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())

				resultMap, ok := result.(map[string]any)
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(resultMap["id"]).To(gomega.Equal("cmd-1"))
				gomega.Expect(resultMap["device_name"]).To(gomega.Equal("dev"))
				mockOrm.AssertExpectations(ginkgo.GinkgoT())
			})
		})

		ginkgo.When("getting command by ID not found", func() {
			ginkgo.It("should return error", func() {
				mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
				mockOrm.On("First", mock.Anything, mock.Anything).Return(mockOrm)
				mockOrm.On("Error").Return(sql.ErrRecordNotFound)
				handler.orm = mockOrm

				result, err := handler.GetByID(context.Background(), "cmd-1")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("getting command"))
				mockOrm.AssertExpectations(ginkgo.GinkgoT())
			})
		})
	})

	ginkgo.Context("Update", func() {
		var (
			orm     *MockORM
			handler *CommandHandler
			mockOrm *MockORM
			cmd     struct {
				ID         string
				DeviceName string
				DeviceID   string
				TaskID     string
			}
		)

		ginkgo.BeforeEach(func() {
			orm = &MockORM{}
			handler = NewCommandHandler(orm)
			mockOrm = &MockORM{}
			cmd = struct {
				ID         string
				DeviceName string
				DeviceID   string
				TaskID     string
			}{ID: "cmd-1", DeviceName: "dev", DeviceID: "dev-1", TaskID: "task-1"}
		})

		ginkgo.When("updating command successfully", func() {
			ginkgo.It("should update command without error", func() {
				mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
				mockOrm.On("Model", mock.Anything).Return(mockOrm)
				mockOrm.On("Updates", mock.Anything).Return(mockOrm)
				mockOrm.On("Error").Return(nil)
				handler.orm = mockOrm

				err := handler.Update(context.Background(), "cmd-1", cmd)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				mockOrm.AssertExpectations(ginkgo.GinkgoT())
			})
		})

		ginkgo.When("updating command with database error", func() {
			ginkgo.It("should return error", func() {
				mockOrm.On("WithContext", mock.Anything).Return(mockOrm)
				mockOrm.On("Model", mock.Anything).Return(mockOrm)
				mockOrm.On("Updates", mock.Anything).Return(mockOrm)
				mockOrm.On("Error").Return(errors.New("db error"))
				handler.orm = mockOrm

				err := handler.Update(context.Background(), "cmd-1", cmd)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("updating command"))
				mockOrm.AssertExpectations(ginkgo.GinkgoT())
			})
		})
	})

	ginkgo.Context("ExtractCommandFields", func() {
		var handler *CommandHandler

		ginkgo.BeforeEach(func() {
			orm := &MockORM{}
			handler = NewCommandHandler(orm)
		})

		ginkgo.It("should extract command fields correctly", func() {
			cmd := struct {
				ID         string
				DeviceName string
				DeviceID   string
				TaskID     string
				Payload    CommandPayload
			}{
				ID:         "cmd-1",
				DeviceName: "dev",
				DeviceID:   "dev-1",
				TaskID:     "task-1",
				Payload:    CommandPayload{Index: 1, Data: 100},
			}

			fields := handler.extractCommandFields(cmd)
			gomega.Expect(fields).NotTo(gomega.BeNil())
			gomega.Expect(fields.ID).To(gomega.Equal("cmd-1"))
			gomega.Expect(fields.DeviceName).To(gomega.Equal("dev"))
			gomega.Expect(fields.DeviceID).To(gomega.Equal("dev-1"))
			gomega.Expect(fields.TaskID).To(gomega.Equal("task-1"))
		})
	})
})
