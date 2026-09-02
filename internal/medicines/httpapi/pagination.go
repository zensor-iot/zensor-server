package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"zensor-server/internal/infra/httpserver"
	"zensor-server/internal/medicines/usecases"

	shareddomain "zensor-server/internal/shared_kernel/domain"
)

// listPaginated is a copy of the maintenance module's helper. Promoting it to
// infra/httpserver would make that package depend on a usecases level
// Pagination type, a worse dependency than the duplication.
func listPaginated[T, R any](
	w http.ResponseWriter,
	r *http.Request,
	id shareddomain.ID,
	listFunc func(ctx context.Context, id shareddomain.ID, pagination usecases.Pagination) ([]T, int, error),
	toResponse func(T) R,
	listingLabel string,
) {
	paginationParams := httpserver.ExtractPaginationParams(r)
	pagination := usecases.Pagination{
		Limit:  paginationParams.Limit,
		Offset: (paginationParams.Page - 1) * paginationParams.Limit,
	}

	items, total, err := listFunc(r.Context(), id, pagination)
	if err != nil {
		slog.Error("listing "+listingLabel, slog.String("error", err.Error()))
		http.Error(w, "failed to list "+listingLabel, http.StatusInternalServerError)
		return
	}

	responses := make([]R, len(items))
	for i, item := range items {
		responses[i] = toResponse(item)
	}

	httpserver.ReplyWithPaginatedData(w, http.StatusOK, responses, total, paginationParams)
}
