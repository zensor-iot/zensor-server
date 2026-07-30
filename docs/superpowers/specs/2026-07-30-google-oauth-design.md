# Google OAuth with Email Allowlist — Design

**Date:** 2026-07-30
**Status:** Approved

## Problem

Zensor Server currently has no real authentication. Identity comes from trusting
`X-User-ID`, `X-User-Name`, and `X-User-Email` request headers
(`internal/infra/httpserver/server.go`), which any client can spoof. `/v1/me`
echoes them back and the k8s deployment has no auth proxy in front.

## Goal

Implement authentication inside the app: any Google account can attempt to sign
in, but only email addresses explicitly enabled by an admin are let in.

## Decisions

- **AuthN lives in the app** — no external proxy (oauth2-proxy etc.).
- **Web portal only** for this iteration; mobile app auth comes later.
- **Allowlist in Postgres, managed from an admin UI** in the web portal.
- **Non-enabled emails are rejected** at login (no pending/self-registration
  state); admins pre-add emails.
- **Admin role**: allowlist entries carry an `is_admin` flag; the first admin is
  bootstrapped from config at startup.
- **Enforce everywhere**: all `/v1/*` and `/ws/*` routes require a valid
  session. Spoofable `X-User-*` headers are stripped from incoming requests.
  Any client relying on them (mobile app, scripts) breaks until it adopts real
  auth.
- **Flow**: server-side OAuth2 authorization-code flow with Redis-backed
  sessions and an HttpOnly cookie. Client secret never leaves the server;
  sessions are revocable instantly.

## Flow

```
Browser → GET /auth/login      → 302 to accounts.google.com (state cookie set)
Google  → GET /auth/callback   → verify state → exchange code → get email (verified)
                                 → lookup email in allowed_users table
                                 → not found: 302 /ui/access-denied (no session)
                                 → found: create Redis session, set cookie, 302 /ui/
```

## Components

Follows the existing `shared_kernel` layout (mirroring tenant_configuration):

- `internal/shared_kernel/domain/allowed_user.go` —
  `AllowedUser{ID, Email, DisplayName, IsAdmin, LastLoginAt}`
- `internal/shared_kernel/usecases/auth_service.go` — login/callback/logout
  logic, allowlist check, admin CRUD. Depends on interfaces:
  `AllowedUserRepository`, `SessionStore`, `OAuthProvider`.
- `internal/shared_kernel/persistence/allowed_user_repository.go` — GORM
  implementation, `allowed_users` table (email unique, stored lowercase).
- `internal/infra/auth/` — `RedisSessionStore` + `MemorySessionStore`
  (for `ENV=local`), `GoogleOAuthProvider` wrapping `golang.org/x/oauth2`
  (interface so tests can mock it).
- `internal/shared_kernel/httpapi/auth_controller.go` — `/auth/*` routes and
  the admin allowed-users API.
- `internal/infra/httpserver/server.go` — auth middleware replaces the current
  user-header middleware.
- New Wire providers in `cmd/api/wire/`.

## Data Model

Table `allowed_users`:

| Column | Type | Notes |
|---|---|---|
| `id` | uuid | primary key; becomes the stable user ID |
| `email` | text | unique, normalized to lowercase |
| `display_name` | text | captured from Google on login |
| `is_admin` | bool | grants access to admin endpoints/UI |
| `last_login_at` | timestamp | nullable |
| `created_at` / `updated_at` | timestamp | |

Bootstrap: at startup, if `auth.bootstrap_admin_email` is set, upsert that
email as an admin.

## Endpoints

| Route | Access | Behavior |
|---|---|---|
| `GET /auth/login` | public | set state cookie, redirect to Google |
| `GET /auth/callback` | public | validate state, exchange code, allowlist check, create session |
| `POST /auth/logout` | session | delete session, clear cookie |
| `GET /v1/me` | session | returns `{user_id, name, email, is_admin}`; 401 if no session (SPA's login trigger) |
| `GET /v1/admin/allowed-users` | admin | list allowlist |
| `POST /v1/admin/allowed-users` | admin | add `{email, is_admin}` |
| `PUT /v1/admin/allowed-users/{id}` | admin | update (toggle admin) |
| `DELETE /v1/admin/allowed-users/{id}` | admin | remove + delete their sessions |

## Enforcement Middleware

Replaces `createUserHeaderMiddleware`:

- Always strips incoming `X-User-*` headers.
- Valid session cookie → sets `X-User-ID/Name/Email` internally from the
  session so existing controllers keep working unchanged; also sets span
  attributes.
- `/v1/*` and `/ws/*` without a session → 401 JSON. Public: `/auth/*`,
  `/healthz`, `/metrics`, `/ui/*`.
- Admin routes additionally check `is_admin` → 403.
- Instant revocation: sessions indexed per user
  (`auth:sessions_by_user:<id>`); deleting an allowed user deletes their
  sessions immediately.
- Session: 256-bit random ID, key `auth:session:<id>` in Redis, cookie
  `HttpOnly; Secure; SameSite=Lax; Path=/`, TTL from config (default 7 days).
- The cookie flows on WebSocket upgrade, so `/ws/` is covered by the same
  middleware.

## Configuration

```yaml
auth:
  enabled: true
  google:
    client_id: ""
    client_secret: ""        # set via ZENSOR_SERVER_AUTH_GOOGLE_CLIENT_SECRET
    redirect_url: "https://portal.zensor-iot.net/auth/callback"
  session_ttl: "168h"
  bootstrap_admin_email: "sebastian@andinolabs.com"
```

`ENV=local`: fake OAuth provider — `/auth/login` immediately creates an admin
session for `dev@localhost` using the in-memory session store, so
`ENV=local just run` keeps working with zero external dependencies.

## Frontend (React SPA)

- `AuthContext` fetches `/v1/me` on load; on 401 renders a login page with a
  "Sign in with Google" button (`window.location = '/auth/login'`).
- `/ui/access-denied` page for authenticated-but-not-enabled users.
- Logout button in the existing `UserInfo` component.
- New admin page "Users" (visible only when `is_admin`): list/add/remove
  emails, toggle admin.
- Global fetch handling: any 401 sends the user back to the login page.

## Testing

Ginkgo unit tests (mocks via mockgen, no network):

- `AuthService`: allowed login, denied login, state mismatch, unverified
  email, logout, admin CRUD, bootstrap upsert.
- Middleware: 401 on missing/invalid session, header stripping, public paths,
  admin gate (403).
- `allowed_users` repository (in-memory SQLite, as existing repo tests do).
- `MemorySessionStore`.

## Trade-offs & Out of Scope

- **User IDs change**: the new stable user ID is the `allowed_users` UUID.
  Existing `users` rows (tenant associations, push tokens) were keyed by
  whatever `X-User-ID` clients previously sent — those associations will be
  orphaned and need re-associating after rollout.
- Google access/refresh tokens are not stored; they are used once at login to
  fetch identity.
- Out of scope: mobile app auth, self-registration/pending state, roles beyond
  `is_admin`.
