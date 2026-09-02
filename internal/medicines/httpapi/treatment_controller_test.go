package httpapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	medicinesDomain "zensor-server/internal/medicines/domain"
	medicinesHTTPAPI "zensor-server/internal/medicines/httpapi"
	medicinesUsecases "zensor-server/internal/medicines/usecases"

	mockmedicines "zensor-server/test/unit/doubles/medicines/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("TreatmentController", func() {
	var (
		ctrl     *gomock.Controller
		service  *mockmedicines.MockTreatmentService
		router   *http.ServeMux
		recorder *httptest.ResponseRecorder
		startAt  time.Time
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		service = mockmedicines.NewMockTreatmentService(ctrl)
		router = http.NewServeMux()
		medicinesHTTPAPI.NewTreatmentController(service).AddRoutes(router)
		recorder = httptest.NewRecorder()
		startAt = time.Now().Add(1 * time.Hour).UTC().Truncate(time.Second)
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	createBody := func(quantity float64, unit string, every int) string {
		return fmt.Sprintf(`{
			"tenant_id": "tenant-1",
			"patient_id": "patient-1",
			"medicine_name": "Amoxicilina",
			"quantity": %v,
			"unit": %q,
			"schedule": {"start_at": %q, "every": %d, "unit": "hour"},
			"total_doses": 21
		}`, quantity, unit, startAt.Format(time.RFC3339), every)
	}

	ginkgo.Context("POST /v1/medicines/treatments", func() {
		ginkgo.When("the payload is valid", func() {
			ginkgo.It("should create the treatment and return its derived fields", func() {
				service.EXPECT().CreateTreatment(gomock.Any(), gomock.Any()).Return(nil)

				request := httptest.NewRequest(http.MethodPost, "/v1/medicines/treatments",
					strings.NewReader(createBody(15, "drop", 8)))
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusCreated))
				var body map[string]any
				gomega.Expect(json.Unmarshal(recorder.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body["dose_text"]).To(gomega.Equal("15 gotas"))
				gomega.Expect(body["ends_at"]).NotTo(gomega.BeNil())
				gomega.Expect(body["is_active"]).To(gomega.BeTrue())
			})
		})

		ginkgo.When("the dose unit is unknown", func() {
			ginkgo.It("should reply 400 without reaching the service", func() {
				request := httptest.NewRequest(http.MethodPost, "/v1/medicines/treatments",
					strings.NewReader(createBody(15, "sachet", 8)))
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusBadRequest))
			})
		})

		ginkgo.When("the schedule cannot produce doses", func() {
			ginkgo.It("should reply 400 without reaching the service", func() {
				request := httptest.NewRequest(http.MethodPost, "/v1/medicines/treatments",
					strings.NewReader(createBody(15, "drop", 0)))
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusBadRequest))
			})
		})

		ginkgo.When("the patient belongs to another tenant", func() {
			ginkgo.It("should reply 400", func() {
				service.EXPECT().CreateTreatment(gomock.Any(), gomock.Any()).
					Return(medicinesUsecases.ErrTenantMismatch)

				request := httptest.NewRequest(http.MethodPost, "/v1/medicines/treatments",
					strings.NewReader(createBody(15, "drop", 8)))
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusBadRequest))
			})
		})

		ginkgo.When("the body is not valid JSON", func() {
			ginkgo.It("should reply 400", func() {
				request := httptest.NewRequest(http.MethodPost, "/v1/medicines/treatments",
					strings.NewReader("not json"))
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusBadRequest))
			})
		})
	})

	ginkgo.Context("GET /v1/medicines/treatments", func() {
		ginkgo.When("neither scope is given", func() {
			ginkgo.It("should reply 400", func() {
				request := httptest.NewRequest(http.MethodGet, "/v1/medicines/treatments", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusBadRequest))
			})
		})

		ginkgo.When("a patient is given", func() {
			ginkgo.It("should scope the listing to that patient", func() {
				service.EXPECT().
					ListTreatmentsByPatient(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, 0, nil)

				request := httptest.NewRequest(http.MethodGet,
					"/v1/medicines/treatments?patient_id=patient-1", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
			})
		})

		ginkgo.When("only a tenant is given", func() {
			ginkgo.It("should scope the listing to that tenant", func() {
				service.EXPECT().
					ListTreatmentsByTenant(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, 0, nil)

				request := httptest.NewRequest(http.MethodGet,
					"/v1/medicines/treatments?tenant_id=tenant-1", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
			})
		})
	})

	ginkgo.Context("PUT /v1/medicines/treatments/{id}", func() {
		ginkgo.When("only some fields are sent", func() {
			ginkgo.It("should leave the rest untouched", func() {
				stored, err := medicinesDomain.NewTreatmentBuilder().
					WithTenantID("tenant-1").WithPatientID("patient-1").
					WithMedicineName("Amoxicilina").
					WithDose(15, medicinesDomain.DoseUnitDrop).
					WithSchedule(medicinesDomain.MedicineSchedule{
						StartAt: startAt, Every: 8, Unit: medicinesDomain.IntervalUnitHour,
					}).Build()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				service.EXPECT().GetTreatment(gomock.Any(), gomock.Any()).Return(stored, nil)
				service.EXPECT().
					UpdateTreatment(gomock.Any(), gomock.Cond(func(x any) bool {
						updated, ok := x.(medicinesDomain.Treatment)
						return ok && updated.Quantity == 20 &&
							updated.MedicineName == "Amoxicilina" &&
							updated.Schedule.Every == 8
					})).
					Return(nil)

				request := httptest.NewRequest(http.MethodPut, "/v1/medicines/treatments/"+stored.ID.String(),
					strings.NewReader(`{"quantity": 20}`))
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
			})
		})
	})

	ginkgo.Context("POST /v1/medicines/treatments/{id}/deactivate", func() {
		ginkgo.When("the treatment exists", func() {
			ginkgo.It("should reply 204", func() {
				service.EXPECT().DeactivateTreatment(gomock.Any(), gomock.Any()).Return(nil)

				request := httptest.NewRequest(http.MethodPost,
					"/v1/medicines/treatments/treatment-1/deactivate", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusNoContent))
			})
		})

		ginkgo.When("the treatment does not exist", func() {
			ginkgo.It("should reply 404", func() {
				service.EXPECT().DeactivateTreatment(gomock.Any(), gomock.Any()).
					Return(medicinesUsecases.ErrTreatmentNotFound)

				request := httptest.NewRequest(http.MethodPost,
					"/v1/medicines/treatments/missing/deactivate", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusNotFound))
			})
		})
	})
})
