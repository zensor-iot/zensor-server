// Package wire contains the Wire dependency injection providers.
package wire

import (
	"fmt"
	"zensor-server/internal/infra/cache"
	"zensor-server/internal/shared_kernel/httpapi"
	"zensor-server/internal/shared_kernel/persistence"
	"zensor-server/internal/shared_kernel/usecases"
)

// APIKeyComponents bundles everything the server needs for API key authentication.
type APIKeyComponents struct {
	Controller *httpapi.APIKeyController
	Service    usecases.APIKeyService
}

// InitializeAPIKeyComponents wires the API key repository, a dedicated
// in-process cache, and the service.
func InitializeAPIKeyComponents() (*APIKeyComponents, error) {
	appConfig := provideAppConfig()
	orm := provideDatabase(appConfig)

	repository, err := persistence.NewAPIKeyRepository(orm)
	if err != nil {
		return nil, fmt.Errorf("creating api key repository: %w", err)
	}

	keyCache, err := cache.New(cache.DefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("creating api key cache: %w", err)
	}

	service := usecases.NewAPIKeyService(repository, keyCache)

	return &APIKeyComponents{
		Controller: httpapi.NewAPIKeyController(service),
		Service:    service,
	}, nil
}
