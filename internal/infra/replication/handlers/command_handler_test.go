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

var _ = ginkgo.Describe("CommandHandler", func() {
	ginkgo.Context("Create", func() {
		var (
			ctrl    *gomock.Controller
			mockOrm *mocksql.MockORM
			handler *handlers.CommandHandler
			cmd     struct {
				ID         string
				DeviceName string
				DeviceID   string
				TaskID     string
			}
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
			handler = handlers.NewCommandHandler(mockOrm)
			cmd = struct {
				ID         string
				DeviceName string
				DeviceID   string
				TaskID     string
			}{ID: "cmd-1", DeviceName: "dev", DeviceID: "dev-1", TaskID: "task-1"}
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.When("creating command successfully", func() {
			ginkgo.It("should create command without error", func() {
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Create(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(nil)

				err := handler.Create(context.Background(), "cmd-1", cmd)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.When("creating command with database error", func() {
			ginkgo.It("should return error", func() {
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Create(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(errors.New("db error"))

				err := handler.Create(context.Background(), "cmd-1", cmd)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("creating command"))
			})
		})
	})

	ginkgo.Context("GetByID", func() {
		var (
			ctrl    *gomock.Controller
			mockOrm *mocksql.MockORM
			handler *handlers.CommandHandler
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
			handler = handlers.NewCommandHandler(mockOrm)
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.When("getting command by ID successfully", func() {
			ginkgo.It("should return command data", func() {
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().First(gomock.Any(), gomock.Any()).DoAndReturn(
					func(dest interface{}, conds ...interface{}) *mocksql.MockORM {
						commandData := dest.(*handlers.CommandData)
						*commandData = handlers.CommandData{
							ID:         "cmd-1",
							DeviceName: "dev",
							DeviceID:   "device-1",
							TaskID:     "task-1",
							Payload:    handlers.CommandPayload{Index: 1, Data: 100},
						}
						return mockOrm
					},
				)
				mockOrm.EXPECT().Error().Return(nil)

				result, err := handler.GetByID(context.Background(), "cmd-1")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())

				resultMap, ok := result.(map[string]interface{})
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(resultMap["id"]).To(gomega.Equal("cmd-1"))
				gomega.Expect(resultMap["device_name"]).To(gomega.Equal("dev"))
			})
		})

		ginkgo.When("getting command by ID not found", func() {
			ginkgo.It("should return error", func() {
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(sql.ErrRecordNotFound)

				result, err := handler.GetByID(context.Background(), "cmd-1")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("getting command"))
			})
		})
	})

	ginkgo.Context("Update", func() {
		var (
			ctrl    *gomock.Controller
			mockOrm *mocksql.MockORM
			handler *handlers.CommandHandler
			cmd     struct {
				ID         string
				DeviceName string
				DeviceID   string
				TaskID     string
				Payload    struct {
					Index int
					Data  int
				}
			}
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
			handler = handlers.NewCommandHandler(mockOrm)
			cmd = struct {
				ID         string
				DeviceName string
				DeviceID   string
				TaskID     string
				Payload    struct {
					Index int
					Data  int
				}
			}{
				ID:         "cmd-1",
				DeviceName: "dev",
				DeviceID:   "dev-1",
				TaskID:     "task-1",
				Payload: struct {
					Index int
					Data  int
				}{Index: 1, Data: 100},
			}
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.When("updating command successfully", func() {
			ginkgo.It("should update command without error", func() {
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm).Times(2)
				mockOrm.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(nil)
				mockOrm.EXPECT().Save(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(nil)

				err := handler.Update(context.Background(), "cmd-1", cmd)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.When("updating command with database error", func() {
			ginkgo.It("should return error", func() {
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(errors.New("db error"))

				err := handler.Update(context.Background(), "cmd-1", cmd)
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("fetching existing command"))
			})
		})
	})
})
