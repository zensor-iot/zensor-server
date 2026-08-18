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

var controllerTestSchedule = maintenanceDomain.Schedule{
	StartDate: time.Now().AddDate(0, 0, 1),
	Every:     1,
	Unit:      maintenanceDomain.RecurrenceUnitMonth,
}

var _ = Describe("MaintenanceActivityController", func() {
	var controller *maintenance_httpapi.ActivityController
	var mockService *mockusecases.MockActivityService
	var ctrl *gomock.Controller
	var recorder *httptest.ResponseRecorder
	var request *http.Request

	BeforeEach(func() {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
		ctrl = gomock.NewController(GinkgoT())
		mockService = mockusecases.NewMockActivityService(ctrl)
		controller = maintenance_httpapi.NewActivityController(mockService)
		recorder = httptest.NewRecorder()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("listActivities", func() {
		var router *http.ServeMux
		var tenantID string

		BeforeEach(func() {
			router = http.NewServeMux()
			controller.AddRoutes(router)
			tenantID = "tenant-123"
		})

		When("successful request with default pagination", func() {
			var activities []maintenanceDomain.Activity

			BeforeEach(func() {
				activityType := maintenanceDomain.PredefinedActivityTypes[maintenanceDomain.ActivityTypeWaterSystem]
				activity1, _ := maintenanceDomain.NewActivityBuilder().
					WithTenantID(shareddomain.ID(tenantID)).
					WithType(activityType).
					WithName("Activity 1").
					WithDescription("Description 1").
					WithSchedule(controllerTestSchedule).
					WithFields([]maintenanceDomain.FieldDefinition{}).
					Build()

				activity2, _ := maintenanceDomain.NewActivityBuilder().
					WithTenantID(shareddomain.ID(tenantID)).
					WithType(activityType).
					WithName("Activity 2").
					WithDescription("Description 2").
					WithSchedule(maintenanceDomain.Schedule{
						StartDate: time.Now().AddDate(0, 0, 1),
						Every:     1,
						Unit:      maintenanceDomain.RecurrenceUnitMonth,
					}).
					WithFields([]maintenanceDomain.FieldDefinition{}).
					Build()

				activities = []maintenanceDomain.Activity{activity1, activity2}
				request = httptest.NewRequest(http.MethodGet, "/v1/maintenance/activities?tenant_id="+tenantID, nil)
			})

			It("should return paginated response with default parameters", func() {
				expectedPagination := maintenance_usecases.Pagination{
					Limit:  10,
					Offset: 0,
				}
				mockService.EXPECT().
					ListActivitiesByTenant(gomock.Any(), shareddomain.ID(tenantID), expectedPagination).
					Return(activities, 2, nil)

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response maintenance_httpapi_internal.ActivityListResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.Data).To(HaveLen(2))
			})
		})

		When("tenant_id is missing", func() {
			BeforeEach(func() {
				request = httptest.NewRequest(http.MethodGet, "/v1/maintenance/activities", nil)
			})

			It("should return bad request", func() {
				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
				Expect(recorder.Body.String()).To(ContainSubstring("tenant_id is required"))
			})
		})

		When("service returns error", func() {
			BeforeEach(func() {
				request = httptest.NewRequest(http.MethodGet, "/v1/maintenance/activities?tenant_id="+tenantID, nil)
			})

			It("should return internal server error", func() {
				expectedPagination := maintenance_usecases.Pagination{
					Limit:  10,
					Offset: 0,
				}
				mockService.EXPECT().
					ListActivitiesByTenant(gomock.Any(), shareddomain.ID(tenantID), expectedPagination).
					Return(nil, 0, errors.New("database error"))

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
				Expect(recorder.Body.String()).To(ContainSubstring("failed to list maintenance activities"))
			})
		})
	})

	Context("getActivity", func() {
		var router *http.ServeMux
		var activityID string

		BeforeEach(func() {
			router = http.NewServeMux()
			controller.AddRoutes(router)
			activityID = "activity-123"
		})

		When("activity exists", func() {
			var activity maintenanceDomain.Activity

			BeforeEach(func() {
				activityType := maintenanceDomain.PredefinedActivityTypes[maintenanceDomain.ActivityTypeWaterSystem]
				activity, _ = maintenanceDomain.NewActivityBuilder().
					WithTenantID(shareddomain.ID("tenant-123")).
					WithType(activityType).
					WithName("Test Activity").
					WithDescription("Test Description").
					WithSchedule(controllerTestSchedule).
					WithFields([]maintenanceDomain.FieldDefinition{}).
					Build()

				request = httptest.NewRequest(http.MethodGet, "/v1/maintenance/activities/"+activityID, nil)
			})

			It("should return the activity", func() {
				mockService.EXPECT().
					GetActivity(gomock.Any(), shareddomain.ID(activityID)).
					Return(activity, nil)

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusOK))

				var response maintenance_httpapi_internal.ActivityResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				Expect(err).NotTo(HaveOccurred())
				Expect(response.ID).To(Equal(activity.ID.String()))
				Expect(response.Name).To(Equal("Test Activity"))
			})
		})

		When("activity not found", func() {
			BeforeEach(func() {
				request = httptest.NewRequest(http.MethodGet, "/v1/maintenance/activities/"+activityID, nil)
			})

			It("should return not found", func() {
				mockService.EXPECT().
					GetActivity(gomock.Any(), shareddomain.ID(activityID)).
					Return(maintenanceDomain.Activity{}, maintenance_usecases.ErrActivityNotFound)

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusNotFound))
				Expect(recorder.Body.String()).To(ContainSubstring("maintenance activity not found"))
			})
		})

		When("service returns error", func() {
			BeforeEach(func() {
				request = httptest.NewRequest(http.MethodGet, "/v1/maintenance/activities/"+activityID, nil)
			})

			It("should return internal server error", func() {
				mockService.EXPECT().
					GetActivity(gomock.Any(), shareddomain.ID(activityID)).
					Return(maintenanceDomain.Activity{}, errors.New("database error"))

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Context("createActivity", func() {
		var router *http.ServeMux
		var createRequest maintenance_httpapi_internal.ActivityCreateRequest

		BeforeEach(func() {
			router = http.NewServeMux()
			controller.AddRoutes(router)

			createRequest = maintenance_httpapi_internal.ActivityCreateRequest{
				TenantID:    "tenant-123",
				TypeName:    maintenanceDomain.ActivityTypeWaterSystem,
				Name:        "Test Activity",
				Description: "Test Description",
				Schedule: maintenance_httpapi_internal.ScheduleRequest{
					StartDate: time.Now().AddDate(0, 0, 1),
					Every:     1,
					Unit:      string(maintenanceDomain.RecurrenceUnitMonth),
				},
				NotificationDaysBefore: []int{7, 3},
				Fields:                 []maintenance_httpapi_internal.FieldDefinitionRequest{},
			}
		})

		When("valid request with predefined type", func() {
			It("should create activity successfully", func() {
				mockService.EXPECT().
					CreateActivity(gomock.Any(), gomock.Any()).
					Return(nil)

				body, _ := json.Marshal(createRequest)
				request = httptest.NewRequest(http.MethodPost, "/v1/maintenance/activities", bytes.NewReader(body))
				request.Header.Set("Content-Type", "application/json")

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusCreated))
			})
		})

		When("invalid JSON body", func() {
			It("should return bad request", func() {
				request = httptest.NewRequest(http.MethodPost, "/v1/maintenance/activities", bytes.NewReader([]byte("invalid json")))
				request.Header.Set("Content-Type", "application/json")

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusBadRequest))
			})
		})

		When("service returns error", func() {
			It("should return internal server error", func() {
				mockService.EXPECT().
					CreateActivity(gomock.Any(), gomock.Any()).
					Return(errors.New("database error"))

				body, _ := json.Marshal(createRequest)
				request = httptest.NewRequest(http.MethodPost, "/v1/maintenance/activities", bytes.NewReader(body))
				request.Header.Set("Content-Type", "application/json")

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})

	Context("updateActivity", func() {
		var router *http.ServeMux
		var activityID string
		var activity maintenanceDomain.Activity
		var updateRequest maintenance_httpapi_internal.ActivityUpdateRequest

		BeforeEach(func() {
			router = http.NewServeMux()
			controller.AddRoutes(router)
			activityID = "activity-123"

			activityType := maintenanceDomain.PredefinedActivityTypes[maintenanceDomain.ActivityTypeWaterSystem]
			activity, _ = maintenanceDomain.NewActivityBuilder().
				WithTenantID(shareddomain.ID("tenant-123")).
				WithType(activityType).
				WithName("Original Name").
				WithDescription("Original Description").
				WithSchedule(controllerTestSchedule).
				WithFields([]maintenanceDomain.FieldDefinition{}).
				Build()

			updateRequest = maintenance_httpapi_internal.ActivityUpdateRequest{
				Name: stringPtr("Updated Name"),
			}
		})

		When("activity exists and update is valid", func() {
			It("should update activity successfully", func() {
				mockService.EXPECT().
					GetActivity(gomock.Any(), shareddomain.ID(activityID)).
					Return(activity, nil)
				mockService.EXPECT().
					UpdateActivity(gomock.Any(), gomock.Any()).
					Return(nil)

				body, _ := json.Marshal(updateRequest)
				request = httptest.NewRequest(http.MethodPut, "/v1/maintenance/activities/"+activityID, bytes.NewReader(body))
				request.Header.Set("Content-Type", "application/json")

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusOK))
			})
		})

		When("activity not found", func() {
			It("should return not found", func() {
				mockService.EXPECT().
					GetActivity(gomock.Any(), shareddomain.ID(activityID)).
					Return(maintenanceDomain.Activity{}, maintenance_usecases.ErrActivityNotFound)

				body, _ := json.Marshal(updateRequest)
				request = httptest.NewRequest(http.MethodPut, "/v1/maintenance/activities/"+activityID, bytes.NewReader(body))
				request.Header.Set("Content-Type", "application/json")

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusNotFound))
			})
		})
	})

	Context("deleteActivity", func() {
		var router *http.ServeMux
		var activityID string

		BeforeEach(func() {
			router = http.NewServeMux()
			controller.AddRoutes(router)
			activityID = "activity-123"
		})

		When("activity exists", func() {
			It("should delete activity successfully", func() {
				mockService.EXPECT().
					DeleteActivity(gomock.Any(), shareddomain.ID(activityID)).
					Return(nil)

				request = httptest.NewRequest(http.MethodDelete, "/v1/maintenance/activities/"+activityID, nil)

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusNoContent))
			})
		})

		When("activity not found", func() {
			It("should return not found", func() {
				mockService.EXPECT().
					DeleteActivity(gomock.Any(), shareddomain.ID(activityID)).
					Return(maintenance_usecases.ErrActivityNotFound)

				request = httptest.NewRequest(http.MethodDelete, "/v1/maintenance/activities/"+activityID, nil)

				router.ServeHTTP(recorder, request)

				Expect(recorder.Code).To(Equal(http.StatusNotFound))
			})
		})
	})

	registerActivityStateContext := func(action, successDescription string, setup func(id shareddomain.ID) *gomock.Call) {
		Context(action, func() {
			var router *http.ServeMux
			var activityID string

			BeforeEach(func() {
				router = http.NewServeMux()
				controller.AddRoutes(router)
				activityID = "activity-123"
			})

			When("activity exists", func() {
				It(successDescription, func() {
					setup(shareddomain.ID(activityID)).Return(nil)

					request = httptest.NewRequest(http.MethodPost, "/v1/maintenance/activities/"+activityID+"/"+action, nil)

					router.ServeHTTP(recorder, request)

					Expect(recorder.Code).To(Equal(http.StatusOK))
				})
			})

			When("activity not found", func() {
				It("should return not found", func() {
					setup(shareddomain.ID(activityID)).Return(maintenance_usecases.ErrActivityNotFound)

					request = httptest.NewRequest(http.MethodPost, "/v1/maintenance/activities/"+activityID+"/"+action, nil)

					router.ServeHTTP(recorder, request)

					Expect(recorder.Code).To(Equal(http.StatusNotFound))
				})
			})
		})
	}
	registerActivityStateContext("activate", "should activate activity successfully", func(id shareddomain.ID) *gomock.Call {
		return mockService.EXPECT().ActivateActivity(gomock.Any(), id)
	})
	registerActivityStateContext("deactivate", "should deactivate activity successfully", func(id shareddomain.ID) *gomock.Call {
		return mockService.EXPECT().DeactivateActivity(gomock.Any(), id)
	})
})

func stringPtr(s string) *string {
	return &s
}
