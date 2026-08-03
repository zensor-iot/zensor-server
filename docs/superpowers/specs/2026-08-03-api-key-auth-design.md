# API Key Authentication — Design

**Date:** 2026-08-03
**Status:** Approved

## Purpose

Allow non-human clients (other applications, integrations, scripts) to authenticate against the Zensor Server API without a browser session. API keys are stored in PostgreSQL, created and revoked exclusively by admins, and validated through an in-memory ristretto cache with a 24-hour TTL.

## Decisions

| Decision | Choice |
|---|---|
| Presentation | `Authorization: Bearer <key>` header |
| Access scope | Full API except `/v1/admin/*` (machine identities are never admin) |
| Storage | SHA-256 hash only; plaintext shown once at creation |
| Key lifetime | No expiry; revocable by admin at any time |
| Cache | Existing `internal/infra/cache` ristretto implementation, 24h TTL |

## 1. Domain model

New file `internal/shared_kernel/domain/api_key.go`:

```go
type APIKey struct {
    ID        ID
    Name      string    // unique human-readable label, e.g. "grafana-sync"
    KeyHash   string    // SHA-256 hex of the full plaintext key (lookup column)
    KeyPrefix string    // first 12 chars of plaintext, e.g. "zsk_ab12cd34", for identification in listings
    CreatedBy ID        // admin user who created it
    CreatedAt time.Time
}
```

- Key format: `zsk_` + 32 random bytes hex-encoded (~68 chars total), generated with `crypto/rand`.
- Constructed via `NewAPIKeyBuilder()`, following the `AllowedUser` builder pattern.
- Name is required and unique; validation lives in the builder.

**Deliberately excluded (YAGNI):** `LastUsedAt` (inaccurate with a 24h cache and adds a write per request), key expiry, per-key scopes, tenant binding.

## 2. Service and cache

New file `internal/shared_kernel/usecases/api_key_service.go` (`SimpleAPIKeyService`), depending on `APIKeyRepository` (defined in usecases) and the existing `cache.Cache` interface:

- `Create(ctx, name string, createdBy domain.ID) (domain.APIKey, string, error)` — generates the key, stores only the hash, returns the domain object plus the plaintext key exactly once. Duplicate name returns `ErrAPIKeyDuplicated`.
- `Validate(ctx, rawKey string) (domain.APIKey, error)` — computes SHA-256 of `rawKey`, then `cache.GetOrSet("apikey:"+hash, 24*time.Hour, loader)` where the loader calls `repository.GetByHash`. Unknown keys are **not** negatively cached; each miss hits the DB (acceptable at machine-client volumes). Unknown key returns `ErrAPIKeyNotFound`.
- `List(ctx) ([]domain.APIKey, error)`.
- `Revoke(ctx, id domain.ID) error` — deletes the DB row **and** the cache entry (`cache.Delete`), making revocation immediate. Single-instance deployment with an in-process cache means there is no cross-instance invalidation gap.

Cache TTL is a constant: `apiKeyCacheTTL = 24 * time.Hour`.

## 3. Repository

- GORM entity `internal/shared_kernel/persistence/internal/api_key.go` with `FromAPIKey`/`ToDomain` mapping; unique indexes on `key_hash` and `name`.
- `internal/shared_kernel/persistence/api_key_repository.go` (`SimpleAPIKeyRepository`): constructor runs `AutoMigrate`, methods `Create`, `GetByHash`, `GetByID`, `FindAll`, `Delete` — same shape as `SimpleAllowedUserRepository`.

## 4. Middleware

`internal/infra/httpserver/auth.go` changes:

```go
type APIKeyResolver interface {
    Validate(ctx context.Context, rawKey string) (domain.APIKey, error)
}

func NewAuthMiddleware(sessions SessionResolver, apiKeys APIKeyResolver) func(http.Handler) http.Handler
```

Resolution order per request:

1. Session cookie (unchanged behavior for humans).
2. Otherwise `Authorization: Bearer zsk_...` → `apiKeys.Validate`. On success, synthetic identity headers are set: `X-User-ID` = key ID, `X-User-Name` = key name, `X-User-Email` = empty string. Span attributes are set the same way sessions do.
3. Requests authenticated via API key receive **403 on `/v1/admin/*` unconditionally**.
4. Any validation failure (malformed header, unknown key, repository error) → 401, identical to an anonymous request. Repository errors are logged with `slog` but still fail closed as 401.

`NewServerWithAuth` in `server.go` gains the `APIKeyResolver` parameter.

## 5. HTTP API

New file `internal/shared_kernel/httpapi/api_key_controller.go`, routes admin-gated by the existing middleware:

| Route | Behavior |
|---|---|
| `POST /v1/admin/api-keys` | Body `{"name": "..."}`. 201 with `{id, name, key, key_prefix, created_at}` — `key` (plaintext) appears only in this response. 400 on missing/blank name. 409 on duplicate name. |
| `GET /v1/admin/api-keys` | 200 with array of `{id, name, key_prefix, created_at}` — never key material. |
| `DELETE /v1/admin/api-keys/{id}` | 204 on success, 404 if not found. |

`docs/openapi.yaml` is updated with the three endpoints and the bearer security scheme.

## 6. Wiring

- New Wire set (`APIKeySet`) in `cmd/api/wire/` providing repository, a dedicated `cache.New(cache.DefaultConfig())` instance, and the service; `just wire` regenerated.
- The service is passed to `NewServerWithAuth` as the `APIKeyResolver` and to the new controller.
- `ENV=local` works unchanged: the repository runs on the in-memory SQLite ORM.

## 7. Error handling summary

| Condition | Result |
|---|---|
| No credentials on protected route | 401 |
| Malformed/unknown/revoked bearer key | 401 |
| Repository error during validation | 401 (logged) |
| API key on `/v1/admin/*` | 403 |
| Duplicate key name on create | 409 |
| Blank name on create | 400 |
| Revoke unknown ID | 404 |

## 8. Testing

TDD with Ginkgo v2 / gomega throughout (`_test` package suffix, exported methods only):

- **Domain:** builder validation (name required, hash/prefix derivation).
- **Service:** mock `APIKeyRepository` (mockgen, `test/unit/doubles/shared_kernel/usecases/`) + real `RistrettoCache`; covers create/plaintext-once, validate cache hit vs DB fallback, revoke purging cache.
- **Middleware:** extend `auth_test.go` — bearer accepted, admin path 403, precedence of session cookie, fail-closed cases.
- **Controller:** handler tests following `auth_controller_test.go` patterns; mock service.
- Functional tests: none (no existing module covers auth; out of scope).
