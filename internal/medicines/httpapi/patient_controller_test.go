package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	medicinesDomain "zensor-server/internal/medicines/domain"
	medicinesHTTPAPI "zensor-server/internal/medicines/httpapi"
	medicinesUsecases "zensor-server/internal/medicines/usecases"
	sharedUsecases "zensor-server/internal/shared_kernel/usecases"

	mockmedicines "zensor-server/test/unit/doubles/medicines/usecases"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("PatientController", func() {
	var (
		ctrl     *gomock.Controller
		service  *mockmedicines.MockPatientService
		router   *http.ServeMux
		recorder *httptest.ResponseRecorder
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		service = mockmedicines.NewMockPatientService(ctrl)
		router = http.NewServeMux()
		medicinesHTTPAPI.NewPatientController(service).AddRoutes(router)
		recorder = httptest.NewRecorder()
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Context("POST /v1/medicines/patients", func() {
		ginkgo.When("the payload is valid", func() {
			ginkgo.It("should create an animal patient", func() {
				service.EXPECT().CreatePatient(gomock.Any(), gomock.Any()).Return(nil)

				request := httptest.NewRequest(http.MethodPost, "/v1/medicines/patients",
					strings.NewReader(`{"tenant_id":"tenant-1","name":"Luna","kind":"animal","notes":"perra"}`))
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusCreated))
				var body map[string]any
				gomega.Expect(json.Unmarshal(recorder.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body["kind"]).To(gomega.Equal("animal"))
			})
		})

		ginkgo.When("the kind is unknown", func() {
			ginkgo.It("should reply 400 without reaching the service", func() {
				request := httptest.NewRequest(http.MethodPost, "/v1/medicines/patients",
					strings.NewReader(`{"tenant_id":"tenant-1","name":"Luna","kind":"plant"}`))
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusBadRequest))
			})
		})

		ginkgo.When("the tenant does not exist", func() {
			ginkgo.It("should reply 400", func() {
				service.EXPECT().CreatePatient(gomock.Any(), gomock.Any()).
					Return(sharedUsecases.ErrTenantNotFound)

				request := httptest.NewRequest(http.MethodPost, "/v1/medicines/patients",
					strings.NewReader(`{"tenant_id":"missing","name":"Luna","kind":"human"}`))
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusBadRequest))
			})
		})
	})

	ginkgo.Context("GET /v1/medicines/patients", func() {
		ginkgo.When("the tenant is missing", func() {
			ginkgo.It("should reply 400", func() {
				request := httptest.NewRequest(http.MethodGet, "/v1/medicines/patients", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusBadRequest))
			})
		})

		ginkgo.When("the tenant is given", func() {
			ginkgo.It("should return the listing", func() {
				service.EXPECT().ListPatientsByTenant(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, 0, nil)

				request := httptest.NewRequest(http.MethodGet, "/v1/medicines/patients?tenant_id=tenant-1", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
			})
		})
	})

	ginkgo.Context("DELETE /v1/medicines/patients/{id}", func() {
		ginkgo.When("the patient still has active treatments", func() {
			ginkgo.It("should reply 409", func() {
				service.EXPECT().DeletePatient(gomock.Any(), gomock.Any()).
					Return(medicinesUsecases.ErrPatientHasActiveTreatments)

				request := httptest.NewRequest(http.MethodDelete, "/v1/medicines/patients/patient-1", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusConflict))
			})
		})

		ginkgo.When("the patient can be removed", func() {
			ginkgo.It("should reply 204", func() {
				service.EXPECT().DeletePatient(gomock.Any(), gomock.Any()).Return(nil)

				request := httptest.NewRequest(http.MethodDelete, "/v1/medicines/patients/patient-1", nil)
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusNoContent))
			})
		})
	})

	ginkgo.Context("PUT /v1/medicines/patients/{id}", func() {
		ginkgo.When("only the notes are sent", func() {
			ginkgo.It("should leave the name and kind untouched", func() {
				stored, err := medicinesDomain.NewPatientBuilder().
					WithTenantID("tenant-1").WithName("Luna").
					WithKind(medicinesDomain.PatientKindAnimal).Build()
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				service.EXPECT().GetPatient(gomock.Any(), gomock.Any()).Return(stored, nil)
				service.EXPECT().
					UpdatePatient(gomock.Any(), gomock.Cond(func(x any) bool {
						updated, ok := x.(medicinesDomain.Patient)
						return ok && updated.Notes == "14 kg" &&
							updated.Name == "Luna" &&
							updated.Kind == medicinesDomain.PatientKindAnimal
					})).
					Return(nil)

				request := httptest.NewRequest(http.MethodPut, "/v1/medicines/patients/"+stored.ID.String(),
					strings.NewReader(`{"notes":"14 kg"}`))
				router.ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
			})
		})
	})
})
