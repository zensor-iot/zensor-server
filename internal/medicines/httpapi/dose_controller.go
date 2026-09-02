package httpapi

import (
	"net/http"
	"time"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/medicines/httpapi/internal"
	"zensor-server/internal/medicines/usecases"

	medicinesDomain "zensor-server/internal/medicines/domain"

	shareddomain "zensor-server/internal/shared_kernel/domain"
)

const (
	getDoseErrMessage        = "failed to get dose"
	administerDoseErrMessage = "failed to administer dose"
	skipDoseErrMessage       = "failed to skip dose"
	listAgendaErrMessage     = "failed to list agenda"
	invalidDoseBody          = "invalid dose payload"
	treatmentIDRequired      = "treatment_id is required"
	invalidAgendaWindow      = "from and to must be RFC3339 timestamps"

	_defaultAgendaWindow = 24 * time.Hour
)

func NewDoseController(service usecases.DoseService) *DoseController {
	return &DoseController{service: service}
}

var _ httpserver.Controller = &DoseController{}

type DoseController struct {
	service usecases.DoseService
}

func (c *DoseController) AddRoutes(router *http.ServeMux) {
	router.Handle("GET /v1/medicines/doses", c.listDoses())
	router.Handle("GET /v1/medicines/doses/{id}", c.getDose())
	router.Handle("POST /v1/medicines/doses/{id}/administer", c.administerDose())
	router.Handle("POST /v1/medicines/doses/{id}/skip", c.skipDose())
	router.Handle("GET /v1/medicines/agenda", c.listAgenda())
}

func (c *DoseController) listDoses() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		treatmentID := r.URL.Query().Get("treatment_id")
		if treatmentID == "" {
			http.Error(w, treatmentIDRequired, http.StatusBadRequest)
			return
		}

		listPaginated(w, r, shareddomain.ID(treatmentID),
			c.service.ListDosesByTreatment, internal.ToDoseResponse, "doses")
	}
}

func (c *DoseController) getDose() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dose, err := c.service.GetDose(r.Context(), shareddomain.ID(r.PathValue("id")))
		if err != nil {
			replyWithError(w, err, getDoseErrMessage)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, internal.ToDoseResponse(dose))
	}
}

func (c *DoseController) administerDose() http.HandlerFunc {
	return c.resolveDose(administerDoseErrMessage, func(r *http.Request, id shareddomain.ID, by medicinesDomain.ResolvedBy, notes *medicinesDomain.DoseNotes) error {
		return c.service.AdministerDose(r.Context(), id, by, notes)
	})
}

func (c *DoseController) skipDose() http.HandlerFunc {
	return c.resolveDose(skipDoseErrMessage, func(r *http.Request, id shareddomain.ID, by medicinesDomain.ResolvedBy, notes *medicinesDomain.DoseNotes) error {
		return c.service.SkipDose(r.Context(), id, by, notes)
	})
}

func (c *DoseController) resolveDose(
	errMessage string,
	resolve func(*http.Request, shareddomain.ID, medicinesDomain.ResolvedBy, *medicinesDomain.DoseNotes) error,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body internal.DoseResolveRequest
		if r.ContentLength > 0 {
			if err := httpserver.DecodeJSONBody(r, &body); err != nil {
				http.Error(w, invalidDoseBody, http.StatusBadRequest)
				return
			}
		}

		var notes *medicinesDomain.DoseNotes
		if body.Notes != nil {
			value := medicinesDomain.DoseNotes(*body.Notes)
			notes = &value
		}

		id := shareddomain.ID(r.PathValue("id"))
		if err := resolve(r, id, resolvedBy(r), notes); err != nil {
			replyWithError(w, err, errMessage)
			return
		}

		dose, err := c.service.GetDose(r.Context(), id)
		if err != nil {
			replyWithError(w, err, errMessage)
			return
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, internal.ToDoseResponse(dose))
	}
}

// resolvedBy reads the actor from the headers the auth middleware populates,
// rather than trusting the request body, so the record of who gave a dose is
// auditable.
func resolvedBy(r *http.Request) medicinesDomain.ResolvedBy {
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return medicinesDomain.ResolvedBy(userID)
	}

	return medicinesDomain.ResolvedBy(r.Header.Get("X-User-Email"))
}

func (c *DoseController) listAgenda() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.URL.Query().Get("tenant_id")
		if tenantID == "" {
			http.Error(w, tenantIDRequired, http.StatusBadRequest)
			return
		}

		from, to, err := agendaWindow(r)
		if err != nil {
			http.Error(w, invalidAgendaWindow, http.StatusBadRequest)
			return
		}

		entries, err := c.service.ListAgenda(r.Context(), shareddomain.ID(tenantID), from, to)
		if err != nil {
			replyWithError(w, err, listAgendaErrMessage)
			return
		}

		responses := make([]internal.AgendaEntryResponse, len(entries))
		for i, entry := range entries {
			responses[i] = internal.ToAgendaEntryResponse(entry)
		}

		httpserver.ReplyJSONResponse(w, http.StatusOK, internal.AgendaResponse{Data: responses})
	}
}

func agendaWindow(r *http.Request) (time.Time, time.Time, error) {
	from := time.Now()
	if raw := r.URL.Query().Get("from"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		from = parsed
	}

	to := from.Add(_defaultAgendaWindow)
	if raw := r.URL.Query().Get("to"); raw != "" {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		to = parsed
	}

	return from, to, nil
}
