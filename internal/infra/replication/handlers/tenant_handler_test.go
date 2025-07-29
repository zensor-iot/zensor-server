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

var _ = ginkgo.Describe("TenantHandler", func() {
	ginkgo.Context("Create", func() {
		var (
			ctrl       *gomock.Controller
			mockOrm    *mocksql.MockORM
			handler    *handlers.TenantHandler
			testTenant struct {
				ID    string
				Name  string
				Email string
			}
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
			handler = handlers.NewTenantHandler(mockOrm)
			testTenant = struct {
				ID    string
				Name  string
				Email string
			}{ID: "tenant-1", Name: "Tenant", Email: "tenant@example.com"}
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.When("creating tenant successfully", func() {
			ginkgo.It("should create tenant without error", func() {
				// Set up mock expectations
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Create(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(nil)

				// Execute the method
				err := handler.Create(context.Background(), "tenant-1", testTenant)

				// Assertions
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.When("creating tenant fails", func() {
			ginkgo.It("should return error when database fails", func() {
				// Set up mock expectations
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Create(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(errors.New("db error"))

				// Execute the method
				err := handler.Create(context.Background(), "tenant-1", testTenant)

				// Assertions
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("creating tenant"))
			})
		})
	})

	ginkgo.Context("GetByID", func() {
		var (
			ctrl    *gomock.Controller
			mockOrm *mocksql.MockORM
			handler *handlers.TenantHandler
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
			handler = handlers.NewTenantHandler(mockOrm)
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.When("getting tenant successfully", func() {
			ginkgo.It("should return tenant data", func() {
				// Set up mock expectations
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().First(gomock.Any(), gomock.Any()).DoAndReturn(
					func(dest interface{}, conds ...interface{}) *mocksql.MockORM {
						// Set the destination with test data
						tenantData := dest.(*handlers.TenantData)
						*tenantData = handlers.TenantData{
							ID:    "tenant-1",
							Name:  "Tenant",
							Email: "tenant@example.com",
						}
						return mockOrm
					},
				)
				mockOrm.EXPECT().Error().Return(nil)

				// Execute the method
				result, err := handler.GetByID(context.Background(), "tenant-1")

				// Assertions
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())

				resultMap, ok := result.(map[string]interface{})
				gomega.Expect(ok).To(gomega.BeTrue())
				gomega.Expect(resultMap["id"]).To(gomega.Equal("tenant-1"))
				gomega.Expect(resultMap["name"]).To(gomega.Equal("Tenant"))
			})
		})

		ginkgo.When("tenant not found", func() {
			ginkgo.It("should return error when tenant not found", func() {
				// Set up mock expectations
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(sql.ErrRecordNotFound)

				// Execute the method
				result, err := handler.GetByID(context.Background(), "tenant-1")

				// Assertions
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(result).To(gomega.BeNil())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("getting tenant"))
			})
		})
	})

	ginkgo.Context("Update", func() {
		var (
			ctrl       *gomock.Controller
			mockOrm    *mocksql.MockORM
			handler    *handlers.TenantHandler
			testTenant struct {
				ID    string
				Name  string
				Email string
			}
		)

		ginkgo.BeforeEach(func() {
			ctrl = gomock.NewController(ginkgo.GinkgoT())
			mockOrm = mocksql.NewMockORM(ctrl)
			handler = handlers.NewTenantHandler(mockOrm)
			testTenant = struct {
				ID    string
				Name  string
				Email string
			}{ID: "tenant-1", Name: "Updated Tenant", Email: "updated@example.com"}
		})

		ginkgo.AfterEach(func() {
			ctrl.Finish()
		})

		ginkgo.When("updating tenant successfully", func() {
			ginkgo.It("should update tenant without error", func() {
				// Set up mock expectations - WithContext is called twice in Update method
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm).Times(2)
				mockOrm.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(nil)
				mockOrm.EXPECT().Save(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(nil)

				// Execute the method
				err := handler.Update(context.Background(), "tenant-1", testTenant)

				// Assertions
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})
		})

		ginkgo.When("updating tenant fails", func() {
			ginkgo.It("should return error when database fails", func() {
				// Set up mock expectations - WithContext is called once before First
				mockOrm.EXPECT().WithContext(gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockOrm)
				mockOrm.EXPECT().Error().Return(errors.New("db error"))

				// Execute the method
				err := handler.Update(context.Background(), "tenant-1", testTenant)

				// Assertions
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("fetching existing tenant"))
			})
		})
	})

})
