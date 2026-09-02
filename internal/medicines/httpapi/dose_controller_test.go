package httpapi_test

import (
	"encoding/json"
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

var _ = ginkgo.Describe("DoseController", func() {
	var (
		ctrl     *gomock.Controller
		service  *mockmedicines.MockDoseService
		router   *http.ServeMux
		recorder *httptest.ResponseRecorder
		dose     medicinesDomain.Dose
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		service = mockmedicines.NewMockDoseService(ctrl)
		router = http.NewServeMux()
		medicinesHTTPAPI.NewDoseController(service).AddRoutes(router)
		recorder = httptest.NewRecorder()

		var err error
		dose, err = medicinesDomain.NewDoseBuilder().
			WithTreatmentID("treatment-1").
			WithSequenceNumber(0).
			WithScheduledAt(time.Now().Add(-5*time.Minute)).
			WithDose(15, medicinesDomain.DoseUnitDrop).
			Build()
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Context("GET /v1/medicines/doses/{id}", func() {
		ginkgo.When("the dose exists", func() {
			ginkgo.It("should return it with its derived overdue flag", func() {
				service.EXPECT().GetDose(gomock.Any(), dose.ID).Return(dose, nil)

				request := httptest.NewRequest(http.MethodGet, "/v1/medicines/doses/"+dose.ID.String(), nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
				var body map[string]any
				gomega.Expect(json.Unmarshal(recorder.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body["is_overdue"]).To(gomega.BeTrue())
				gomega.Expect(body["status"]).To(gomega.Equal("pending"))
			})
		})

		ginkgo.When("the dose does not exist", func() {
			ginkgo.It("should reply 404", func() {
				service.EXPECT().GetDose(gomock.Any(), gomock.Any()).
					Return(medicinesDomain.Dose{}, medicinesUsecases.ErrDoseNotFound)

				request := httptest.NewRequest(http.MethodGet, "/v1/medicines/doses/missing", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusNotFound))
			})
		})
	})

	ginkgo.Context("POST /v1/medicines/doses/{id}/administer", func() {
		ginkgo.When("the request is authenticated", func() {
			ginkgo.It("should record the actor from the headers, not from the body", func() {
				service.EXPECT().
					AdministerDose(gomock.Any(), dose.ID, medicinesDomain.ResolvedBy("user-42"), gomock.Any()).
					Return(nil)
				service.EXPECT().GetDose(gomock.Any(), dose.ID).Return(dose, nil)

				request := httptest.NewRequest(http.MethodPost,
					"/v1/medicines/doses/"+dose.ID.String()+"/administer",
					strings.NewReader(`{"notes":"con comida"}`))
				request.Header.Set("X-User-ID", "user-42")
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
			})
		})

		ginkgo.When("there is no body", func() {
			ginkgo.It("should still administer the dose", func() {
				service.EXPECT().AdministerDose(gomock.Any(), dose.ID, gomock.Any(), gomock.Nil()).Return(nil)
				service.EXPECT().GetDose(gomock.Any(), dose.ID).Return(dose, nil)

				request := httptest.NewRequest(http.MethodPost,
					"/v1/medicines/doses/"+dose.ID.String()+"/administer", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
			})
		})

		ginkgo.When("the dose was already resolved", func() {
			ginkgo.It("should reply 409", func() {
				service.EXPECT().AdministerDose(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(medicinesDomain.ErrDoseAlreadyResolved)

				request := httptest.NewRequest(http.MethodPost,
					"/v1/medicines/doses/"+dose.ID.String()+"/administer", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusConflict))
			})
		})
	})

	ginkgo.Context("POST /v1/medicines/doses/{id}/skip", func() {
		ginkgo.When("the dose is pending", func() {
			ginkgo.It("should skip it", func() {
				service.EXPECT().SkipDose(gomock.Any(), dose.ID, gomock.Any(), gomock.Any()).Return(nil)
				service.EXPECT().GetDose(gomock.Any(), dose.ID).Return(dose, nil)

				request := httptest.NewRequest(http.MethodPost,
					"/v1/medicines/doses/"+dose.ID.String()+"/skip", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
			})
		})
	})

	ginkgo.Context("GET /v1/medicines/agenda", func() {
		ginkgo.When("the tenant is missing", func() {
			ginkgo.It("should reply 400", func() {
				request := httptest.NewRequest(http.MethodGet, "/v1/medicines/agenda", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusBadRequest))
			})
		})

		ginkgo.When("no window is given", func() {
			ginkgo.It("should default to the next day", func() {
				service.EXPECT().
					ListAgenda(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ any, _ any, from, to time.Time) ([]medicinesUsecases.DoseWithContext, error) {
						gomega.Expect(to.Sub(from)).To(gomega.Equal(24 * time.Hour))
						return nil, nil
					})

				request := httptest.NewRequest(http.MethodGet, "/v1/medicines/agenda?tenant_id=tenant-1", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
			})
		})

		ginkgo.When("the window is malformed", func() {
			ginkgo.It("should reply 400", func() {
				request := httptest.NewRequest(http.MethodGet,
					"/v1/medicines/agenda?tenant_id=tenant-1&from=yesterday", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusBadRequest))
			})
		})

		ginkgo.When("the window is too wide", func() {
			ginkgo.It("should reply 400", func() {
				service.EXPECT().ListAgenda(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, medicinesUsecases.ErrAgendaWindowTooLarge)

				request := httptest.NewRequest(http.MethodGet, "/v1/medicines/agenda?tenant_id=tenant-1", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusBadRequest))
			})
		})

		ginkgo.When("entries are returned", func() {
			ginkgo.It("should flatten the patient and medicine onto each entry", func() {
				patient, err := medicinesDomain.NewPatientBuilder().
					WithTenantID("tenant-1").WithName("Luna").
					WithKind(medicinesDomain.PatientKindAnimal).Build()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				treatment, err := medicinesDomain.NewTreatmentBuilder().
					WithTenantID("tenant-1").WithPatientID(patient.ID.String()).
					WithMedicineName("Amoxicilina").
					WithDose(15, medicinesDomain.DoseUnitDrop).
					WithSchedule(medicinesDomain.MedicineSchedule{
						StartAt: time.Now(), Every: 8, Unit: medicinesDomain.IntervalUnitHour,
					}).Build()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				service.EXPECT().ListAgenda(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]medicinesUsecases.DoseWithContext{
						{Dose: dose, Treatment: treatment, Patient: patient},
					}, nil)

				request := httptest.NewRequest(http.MethodGet, "/v1/medicines/agenda?tenant_id=tenant-1", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
				var body struct {
					Data []map[string]any `json:"data"`
				}
				gomega.Expect(json.Unmarshal(recorder.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body.Data).To(gomega.HaveLen(1))
				gomega.Expect(body.Data[0]["patient_name"]).To(gomega.Equal("Luna"))
				gomega.Expect(body.Data[0]["dose_text"]).To(gomega.Equal("15 gotas"))
			})
		})
	})
})
