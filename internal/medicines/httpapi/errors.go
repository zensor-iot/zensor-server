// Package httpapi provides HTTP controllers for the medicines module.
package httpapi

import (
	"errors"
	"log/slog"
	"net/http"
	"zensor-server/internal/medicines/usecases"

	medicinesDomain "zensor-server/internal/medicines/domain"

	sharedUsecases "zensor-server/internal/shared_kernel/usecases"
)

// replyWithError maps a service error onto a status code. Everything the caller
// could have got right is a 4xx; anything else is reported as a server error and
// logged with its cause.
func replyWithError(w http.ResponseWriter, err error, fallbackMessage string) {
	switch {
	case errors.Is(err, usecases.ErrPatientNotFound),
		errors.Is(err, usecases.ErrTreatmentNotFound),
		errors.Is(err, usecases.ErrDoseNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, medicinesDomain.ErrDoseAlreadyResolved),
		errors.Is(err, usecases.ErrPatientHasActiveTreatments),
		errors.Is(err, usecases.ErrPatientDeleted),
		errors.Is(err, usecases.ErrTreatmentDeleted):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, usecases.ErrInvalidTreatmentSchedule),
		errors.Is(err, usecases.ErrTenantMismatch),
		errors.Is(err, usecases.ErrAgendaWindowTooLarge),
		errors.Is(err, sharedUsecases.ErrTenantNotFound):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		slog.Error(fallbackMessage, slog.String("error", err.Error()))
		http.Error(w, fallbackMessage, http.StatusInternalServerError)
	}
}
