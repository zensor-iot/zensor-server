package httpapi_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"
	maintenanceDomain "zensor-server/internal/maintenance/domain"
	maintenance_httpapi "zensor-server/internal/maintenance/httpapi"
	maintenance_httpapi_internal "zensor-server/internal/maintenance/httpapi/internal"
	maintenance_usecases "zensor-server/internal/maintenance/usecases"
	shareddomain "zensor-server/internal/shared_kernel/domain"
	mockusecases "zensor-server/test/unit/doubles/maintenance/usecases"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = Describe("MaintenanceExecutionController", func() {
	var controller *maintenance_httpapi.ExecutionController
	var mockExecutionService *mockusecases.MockExecutionService
	var mockActivityService *mockusecases.MockActivityService
	var ctrl *gomock.Controller
	var recorder *httptest.ResponseRecorder
	var request *http.Request

	BeforeEach(func() {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
		ctrl = gomock.NewController(GinkgoT())
		mockExecutionService = mockusecases.NewMockExecutionService(ctrl)
		mockActivityService = mockusecases.NewMockActivityService(ctrl)
		controller = maintenance_httpapi.NewExecutionController(mockExecutionService, mockActivityService)
		recorder = httptest.NewRecorder()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("listExecutions", func() {
		var router *http.ServeMux
		var activityID string

		BeforeEach(func() {
			router = http.NewServeMux()
			controller.AddRoutes(router)
			activityID = "activity-123"
		})

		When("successful request with default pagination", func() {
			var executions []maintenanceDomain.Execution

			BeforeEach(func() {
				execution1, _ := maintenanceDomain.NewExecutionBuilder().
					WithActivityID(shareddomain.ID(activityID)).
					WithScheduledDate(time.Now().AddDate(0, 0, 30)).
					WithFieldValues(map[string]any{"key": "value1"}).
					Build()

				execution2, _ := maintenanceDomain.NewExecutionBuilder().
					WithActivityID(shareddomain.ID(activityID)).
					WithScheduledDate(time.Now().AddDate(0, 0, 60)).
					WithFieldValues(map[string]any{"key": "value2"}).
					Build()

				executions = []maintenanceDomain.Execution{execution1, execution2}
				request = httptest.NewRequest("GET", "/v1/maintenance/executions?activity_id="+activityID, nil)
			})

			It("should return paginated response with default parameters", func() {
				expectedPagination := maintenance_usecases.Pagination{
					Limit:  10,
					Offset: 0,
				}
				mockExecutionService.EXPECT().
					ListExecutionsByActivity(gomock.Any(), shareddomain.ID(activityID), expectedPagination).
					Return(executions, 2, nil)

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response maintenance_httpapi_internal.ExecutionListResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Data).To(HaveLen(2))
			})
		})

		When("activity_id is missing", func() {
			BeforeEach(func() {
				request = httptest.NewRequest("GET", "/v1/maintenance/executions", nil)
			})

			It("should return bad request", func() {
				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(ContainSubstring("activity_id is required"))
			})
		})

		When("service returns error", func() {
			BeforeEach(func() {
				request = httptest.NewRequest("GET", "/v1/maintenance/executions?activity_id="+activityID, nil)
			})

			It("should return internal server error", func() {
				expectedPagination := maintenance_usecases.Pagination{
					Limit:  10,
					Offset: 0,
				}
				mockExecutionService.EXPECT().
					ListExecutionsByActivity(gomock.Any(), shareddomain.ID(activityID), expectedPagination).
					Return(nil, 0, errors.New("database error"))

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Context("getExecution", func() {
		var router *http.ServeMux
		var executionID string

		BeforeEach(func() {
			router = http.NewServeMux()
			controller.AddRoutes(router)
			executionID = "execution-123"
		})

		When("execution exists", func() {
			var execution maintenanceDomain.Execution

			BeforeEach(func() {
				execution, _ = maintenanceDomain.NewExecutionBuilder().
					WithActivityID(shareddomain.ID("activity-123")).
					WithScheduledDate(time.Now().AddDate(0, 0, 30)).
					WithFieldValues(map[string]any{"maintenance_type": "Filter Replacement"}).
					Build()

				request = httptest.NewRequest("GET", "/v1/maintenance/executions/"+executionID, nil)
			})

			It("should return the execution", func() {
				mockExecutionService.EXPECT().
					GetExecution(gomock.Any(), shareddomain.ID(executionID)).
					Return(execution, nil)

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response maintenance_httpapi_internal.ExecutionResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.ID).To(Equal(execution.ID.String()))
			})
		})

		When("execution not found", func() {
			BeforeEach(func() {
				request = httptest.NewRequest("GET", "/v1/maintenance/executions/"+executionID, nil)
			})

			It("should return not found", func() {
				mockExecutionService.EXPECT().
					GetExecution(gomock.Any(), shareddomain.ID(executionID)).
					Return(maintenanceDomain.Execution{}, maintenance_usecases.ErrExecutionNotFound)

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusNotFound))
				Expect(recorder.Body.String()).To(ContainSubstring("maintenance execution not found"))
			})
		})
	})

	Context("markCompleted", func() {
		var router *http.ServeMux
		var executionID string
		var completeRequest maintenance_httpapi_internal.ExecutionCompleteRequest

		BeforeEach(func() {
			router = http.NewServeMux()
			controller.AddRoutes(router)
			executionID = "execution-123"
			completeRequest = maintenance_httpapi_internal.ExecutionCompleteRequest{
				CompletedBy: "user@example.com",
			}
		})

		When("execution exists", func() {
			It("should mark execution as completed", func() {
				mockExecutionService.EXPECT().
					MarkExecutionCompleted(gomock.Any(), shareddomain.ID(executionID), "user@example.com").
					Return(nil)

				body, _ := json.Marshal(completeRequest)
				request = httptest.NewRequest("POST", "/v1/maintenance/executions/"+executionID+"/complete", bytes.NewReader(body))
				request.Header.Set("Content-Type", "application/json")

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusOK))
			})
		})

		When("execution not found", func() {
			It("should return not found", func() {
				mockExecutionService.EXPECT().
					MarkExecutionCompleted(gomock.Any(), shareddomain.ID(executionID), "user@example.com").
					Return(maintenance_usecases.ErrExecutionNotFound)

				body, _ := json.Marshal(completeRequest)
				request = httptest.NewRequest("POST", "/v1/maintenance/executions/"+executionID+"/complete", bytes.NewReader(body))
				request.Header.Set("Content-Type", "application/json")

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusNotFound))
			})
		})

		When("invalid JSON body", func() {
			It("should return bad request", func() {
				request = httptest.NewRequest("POST", "/v1/maintenance/executions/"+executionID+"/complete", bytes.NewReader([]byte("invalid json")))
				request.Header.Set("Content-Type", "application/json")

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})
	})
})
