package httpapi

import (
	"net/http"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/medicines/httpapi/internal"
	"zensor-server/internal/medicines/usecases"

	medicinesDomain "zensor-server/internal/medicines/domain"

	shareddomain "zensor-server/internal/shared_kernel/domain"
)

const (
	createPatientErrMessage = "failed to create patient"
	getPatientErrMessage    = "failed to get patient"
	updatePatientErrMessage = "failed to update patient"
	deletePatientErrMessage = "failed to delete patient"
	invalidPatientBody      = "invalid patient payload"
	tenantIDRequired        = "tenant_id is required"
)

func NewPatientController(service usecases.PatientService) *PatientController {
	return &PatientController{service: service}
}

var _ httpserver.Controller = &PatientController{}

type PatientController struct {
	service usecases.PatientService
}

func (c *PatientController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /v1/medicines/patients", c.listPatients())
	router.Handle("POST /v1/medicines/patients", c.createPatient())
	router.Handle("GET /v1/medicines/patients/{id}", c.getPatient())
	router.Handle("PUT /v1/medicines/patients/{id}", c.updatePatient())
	router.Handle("DELETE /v1/medicines/patients/{id}", c.deletePatient())
}

func (c *PatientController) listPatients() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.URL.Query().Get("tenant_id")
		if tenantID == "" {
			http.Error(w, tenantIDRequired, http.StatusBadRequest)
			return
		}

		listPaginated(w, r, shareddomain.ID(tenantID),
			c.service.ListPatientsByTenant, internal.ToPatientResponse, "patients")
	}
}

func (c *PatientController) createPatient() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body internal.PatientCreateRequest
		if err := httpserver.DecodeJSONBody(r, &body); err != nil {
			http.Error(w, invalidPatientBody, http.StatusBadRequest)
			return
		}

		patient, err := medicinesDomain.NewPatientBuilder().
			WithTenantID(body.TenantID).
			WithName(body.Name).
			WithKind(medicinesDomain.PatientKind(body.Kind)).
			WithNotes(body.Notes).
			Build()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := c.service.CreatePatient(r.Context(), patient); err != nil {
			replyWithError(w, err, createPatientErrMessage)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusCreated, internal.ToPatientResponse(patient))
	}
}

func (c *PatientController) getPatient() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		patient, err := c.service.GetPatient(r.Context(), shareddomain.ID(r.PathValue("id")))
		if err != nil {
			replyWithError(w, err, getPatientErrMessage)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, internal.ToPatientResponse(patient))
	}
}

func (c *PatientController) updatePatient() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body internal.PatientUpdateRequest
		if err := httpserver.DecodeJSONBody(r, &body); err != nil {
			http.Error(w, invalidPatientBody, http.StatusBadRequest)
			return
		}

		patient, err := c.service.GetPatient(r.Context(), shareddomain.ID(r.PathValue("id")))
		if err != nil {
			replyWithError(w, err, getPatientErrMessage)
			return
		}

		if body.Name != nil {
			patient.Name = shareddomain.Name(*body.Name)
		}
		if body.Kind != nil {
			kind := medicinesDomain.PatientKind(*body.Kind)
			if err := kind.Validate(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			patient.Kind = kind
		}
		if body.Notes != nil {
			patient.Notes = shareddomain.Description(*body.Notes)
		}

		if err := c.service.UpdatePatient(r.Context(), patient); err != nil {
			replyWithError(w, err, updatePatientErrMessage)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, internal.ToPatientResponse(patient))
	}
}

func (c *PatientController) deletePatient() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := c.service.DeletePatient(r.Context(), shareddomain.ID(r.PathValue("id"))); err != nil {
			replyWithError(w, err, deletePatientErrMessage)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
