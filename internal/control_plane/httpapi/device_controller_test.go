package httpapi_test

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"zensor-server/internal/control_plane/httpapi"
	"zensor-server/internal/control_plane/usecases"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/shared_kernel/domain"
	mockusecases "zensor-server/test/unit/doubles/control_plane/usecases"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("DeviceController", func() {
	var controller *httpapi.DeviceController
	var mockService *mockusecases.MockDeviceService
	var ctrl *gomock.Controller
	var recorder *httptest.ResponseRecorder
	var request *http.Request

	BeforeEach(func() {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
		ctrl = gomock.NewController(GinkgoT())
		mockService = mockusecases.NewMockDeviceService(ctrl)
		controller = httpapi.NewDeviceController(mockService)
		recorder = httptest.NewRecorder()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("listDevices", func() {
		var router *http.ServeMux

		BeforeEach(func() {
			router = http.NewServeMux()
			controller.AddRoutes(router)
		})

		When("successful request with default pagination", func() {
			var devices []domain.Device

			BeforeEach(func() {
				devices = []domain.Device{
					{
						ID:          domain.ID("device-1"),
						Name:        "device1",
						DisplayName: "Device 1",
						AppEUI:      "app-eui-1",
						DevEUI:      "dev-eui-1",
						AppKey:      "app-key-1",
					},
					{
						ID:          domain.ID("device-2"),
						Name:        "device2",
						DisplayName: "Device 2",
						AppEUI:      "app-eui-2",
						DevEUI:      "dev-eui-2",
						AppKey:      "app-key-2",
					},
				}

				request = httptest.NewRequest("GET", "/v1/devices", nil)
			})

			It("should return paginated response with default parameters", func() {
				expectedPagination := usecases.Pagination{
					Limit:  10,
					Offset: 0,
				}
				mockService.EXPECT().
					AllDevices(gomock.Any(), expectedPagination).
					Return(devices, 2, nil)

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response httpserver.PaginatedResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Pagination.Page).To(Equal(1))
				Expect(response.Pagination.Limit).To(Equal(10))
				Expect(response.Pagination.Total).To(Equal(2))
				Expect(response.Pagination.TotalPages).To(Equal(1))

				deviceData, ok := response.Data.([]any)
				Expect(ok).To(BeTrue())
				Expect(len(deviceData)).To(Equal(2))
			})
		})

		When("successful request with custom pagination", func() {
			var devices []domain.Device

			BeforeEach(func() {
				devices = []domain.Device{
					{
						ID:          domain.ID("device-1"),
						Name:        "device1",
						DisplayName: "Device 1",
					},
				}

				request = httptest.NewRequest("GET", "/v1/devices?page=2&limit=5", nil)
			})

			It("should return paginated response with custom parameters", func() {
				expectedPagination := usecases.Pagination{
					Limit:  5,
					Offset: 5, // (page 2 - 1) * limit 5
				}
				mockService.EXPECT().
					AllDevices(gomock.Any(), expectedPagination).
					Return(devices, 25, nil)

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response httpserver.PaginatedResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Pagination.Page).To(Equal(2))
				Expect(response.Pagination.Limit).To(Equal(5))
				Expect(response.Pagination.Total).To(Equal(25))
				Expect(response.Pagination.TotalPages).To(Equal(5))

				deviceData, ok := response.Data.([]any)
				Expect(ok).To(BeTrue())
				Expect(len(deviceData)).To(Equal(1))
			})
		})

		When("service returns error", func() {
			BeforeEach(func() {
				request = httptest.NewRequest("GET", "/v1/devices", nil)
			})

			It("should return internal server error", func() {
				expectedPagination := usecases.Pagination{
					Limit:  10,
					Offset: 0,
				}
				mockService.EXPECT().
					AllDevices(gomock.Any(), expectedPagination).
					Return(nil, 0, errors.New("database error"))

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(Equal("service all devices\n"))
			})
		})

		When("invalid pagination parameters", func() {
			var devices []domain.Device

			BeforeEach(func() {
				devices = []domain.Device{
					{
						ID:          domain.ID("device-1"),
						Name:        "device1",
						DisplayName: "Device 1",
					},
				}

				request = httptest.NewRequest("GET", "/v1/devices?page=invalid&limit=abc", nil)
			})

			It("should use default pagination parameters", func() {
				expectedPagination := usecases.Pagination{
					Limit:  10,
					Offset: 0,
				}
				mockService.EXPECT().
					AllDevices(gomock.Any(), expectedPagination).
					Return(devices, 1, nil)

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response httpserver.PaginatedResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Pagination.Page).To(Equal(1))
				Expect(response.Pagination.Limit).To(Equal(10))
			})
		})

		When("limit exceeds maximum", func() {
			var devices []domain.Device

			BeforeEach(func() {
				devices = []domain.Device{
					{
						ID:          domain.ID("device-1"),
						Name:        "device1",
						DisplayName: "Device 1",
					},
				}

				request = httptest.NewRequest("GET", "/v1/devices?limit=200", nil)
			})

			It("should cap limit at maximum value", func() {
				expectedPagination := usecases.Pagination{
					Limit:  100, // Capped at maximum
					Offset: 0,
				}
				mockService.EXPECT().
					AllDevices(gomock.Any(), expectedPagination).
					Return(devices, 1, nil)

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response httpserver.PaginatedResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())

				Expect(response.Pagination.Limit).To(Equal(100)) // Maximum limit
			})
		})
	})
})
