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
	createTreatmentErrMessage     = "failed to create treatment"
	getTreatmentErrMessage        = "failed to get treatment"
	updateTreatmentErrMessage     = "failed to update treatment"
	deleteTreatmentErrMessage     = "failed to delete treatment"
	activateTreatmentErrMessage   = "failed to activate treatment"
	deactivateTreatmentErrMessage = "failed to deactivate treatment"
	invalidTreatmentBody          = "invalid treatment payload"
	treatmentScopeRequired        = "patient_id or tenant_id is required"
)

func NewTreatmentController(service usecases.TreatmentService) *TreatmentController {
	return &TreatmentController{service: service}
}

var _ httpserver.Controller = &TreatmentController{}

type TreatmentController struct {
	service usecases.TreatmentService
}

func (c *TreatmentController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /v1/medicines/treatments", c.listTreatments())
	router.Handle("POST /v1/medicines/treatments", c.createTreatment())
	router.Handle("GET /v1/medicines/treatments/{id}", c.getTreatment())
	router.Handle("PUT /v1/medicines/treatments/{id}", c.updateTreatment())
	router.Handle("DELETE /v1/medicines/treatments/{id}", c.deleteTreatment())
	router.Handle("POST /v1/medicines/treatments/{id}/activate", c.activateTreatment())
	router.Handle("POST /v1/medicines/treatments/{id}/deactivate", c.deactivateTreatment())
}

// listTreatments scopes by patient when a patient_id is given, and by tenant
// otherwise, so the same route serves both the patient detail page and any
// tenant wide listing.
func (c *TreatmentController) listTreatments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if patientID := r.URL.Query().Get("patient_id"); patientID != "" {
			listPaginated(w, r, shareddomain.ID(patientID),
				c.service.ListTreatmentsByPatient, internal.ToTreatmentResponse, "treatments")
			return
		}

		tenantID := r.URL.Query().Get("tenant_id")
		if tenantID == "" {
			http.Error(w, treatmentScopeRequired, http.StatusBadRequest)
			return
		}

		listPaginated(w, r, shareddomain.ID(tenantID),
			c.service.ListTreatmentsByTenant, internal.ToTreatmentResponse, "treatments")
	}
}

func (c *TreatmentController) createTreatment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body internal.TreatmentCreateRequest
		if err := httpserver.DecodeJSONBody(r, &body); err != nil {
			http.Error(w, invalidTreatmentBody, http.StatusBadRequest)
			return
		}

		treatment, err := medicinesDomain.NewTreatmentBuilder().
			WithTenantID(body.TenantID).
			WithPatientID(body.PatientID).
			WithMedicineName(body.MedicineName).
			WithDose(body.Quantity, medicinesDomain.DoseUnit(body.Unit)).
			WithSchedule(internal.ToMedicineSchedule(body.Schedule)).
			WithEndAt(body.EndAt).
			WithTotalDoses(body.TotalDoses).
			WithNotes(body.Notes).
			Build()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := c.service.CreateTreatment(r.Context(), treatment); err != nil {
			replyWithError(w, err, createTreatmentErrMessage)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusCreated, internal.ToTreatmentResponse(treatment))
	}
}

func (c *TreatmentController) getTreatment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		treatment, err := c.service.GetTreatment(r.Context(), shareddomain.ID(r.PathValue("id")))
		if err != nil {
			replyWithError(w, err, getTreatmentErrMessage)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, internal.ToTreatmentResponse(treatment))
	}
}

func (c *TreatmentController) updateTreatment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body internal.TreatmentUpdateRequest
		if err := httpserver.DecodeJSONBody(r, &body); err != nil {
			http.Error(w, invalidTreatmentBody, http.StatusBadRequest)
			return
		}

		treatment, err := c.service.GetTreatment(r.Context(), shareddomain.ID(r.PathValue("id")))
		if err != nil {
			replyWithError(w, err, getTreatmentErrMessage)
			return
		}

		applyTreatmentUpdate(&treatment, body)

		if err := c.service.UpdateTreatment(r.Context(), treatment); err != nil {
			replyWithError(w, err, updateTreatmentErrMessage)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, internal.ToTreatmentResponse(treatment))
	}
}

func applyTreatmentUpdate(treatment *medicinesDomain.Treatment, body internal.TreatmentUpdateRequest) {
	if body.MedicineName != nil {
		treatment.MedicineName = medicinesDomain.MedicineName(*body.MedicineName)
	}
	if body.Quantity != nil {
		treatment.Quantity = medicinesDomain.DoseQuantity(*body.Quantity)
	}
	if body.Unit != nil {
		treatment.Unit = medicinesDomain.DoseUnit(*body.Unit)
	}
	if body.Schedule != nil {
		treatment.Schedule = internal.ToMedicineSchedule(*body.Schedule)
	}
	if body.EndAt != nil {
		treatment.EndAt = body.EndAt
	}
	if body.TotalDoses != nil {
		treatment.TotalDoses = body.TotalDoses
	}
	if body.Notes != nil {
		treatment.Notes = shareddomain.Description(*body.Notes)
	}
}

func (c *TreatmentController) deleteTreatment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := c.service.DeleteTreatment(r.Context(), shareddomain.ID(r.PathValue("id"))); err != nil {
			replyWithError(w, err, deleteTreatmentErrMessage)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *TreatmentController) activateTreatment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := c.service.ActivateTreatment(r.Context(), shareddomain.ID(r.PathValue("id"))); err != nil {
			replyWithError(w, err, activateTreatmentErrMessage)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *TreatmentController) deactivateTreatment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := c.service.DeactivateTreatment(r.Context(), shareddomain.ID(r.PathValue("id"))); err != nil {
			replyWithError(w, err, deactivateTreatmentErrMessage)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
