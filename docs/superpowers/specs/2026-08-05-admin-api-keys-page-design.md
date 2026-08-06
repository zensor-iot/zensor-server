# Admin API Keys Page — Design

**Date:** 2026-08-05
**Status:** Approved

## Purpose

The API key authentication feature ([2026-08-03-api-key-auth-design.md](2026-08-03-api-key-auth-design.md)) shipped backend-only: the `/v1/admin/api-keys` endpoints exist, but nothing in the web UI reaches them. Today an admin can only mint a key with a hand-rolled `curl` or a browser-console `fetch`. This adds an `/admin/api-keys` page so admins can list, create, and revoke keys from the portal.

No backend changes. The endpoints in `internal/shared_kernel/httpapi/api_key_controller.go` are used exactly as they are.

## Decisions

| Decision | Choice |
|---|---|
| Route | `/admin/api-keys`, alongside the other admin routes in `App.jsx` |
| Reveal UX | Blocking modal after creation, dismissible only via an explicit button |
| Network layer | Extracted to `web/src/utils/apiKeys.js` so it is unit-testable without a DOM |
| Testing | Vitest unit tests against the utils module; no new test dependencies |
| Component pattern | Copied from `AdminUsers.jsx` — same header, overlay form, table, and notification usage |

## 1. API contract in use

Already implemented server-side; no changes.

| Method | Path | Notes |
|---|---|---|
| `GET` | `/v1/admin/api-keys` | Returns `[{id, name, key_prefix, created_at}]` |
| `POST` | `/v1/admin/api-keys` | Body `{name}`. `201` returns `{id, name, key, key_prefix, created_at}` — `key` is the plaintext, returned exactly once. `409` duplicate name, `400` missing name |
| `DELETE` | `/v1/admin/api-keys/{id}` | `204` on success, `404` unknown id. Hard-deletes the row and evicts the cache entry, so revocation is immediate |

All three sit behind `adminPathPrefix` in `internal/infra/httpserver/auth.go`, which requires an admin **session cookie** and explicitly rejects bearer-API-key auth. The page therefore works only for a signed-in admin — the same audience as every other `/admin/*` route — and the browser sends the cookie automatically.

## 2. Network module — `web/src/utils/apiKeys.js`

Three functions wrapping `fetch`, each mapping HTTP status onto a typed error so the component never inspects status codes:

```js
export class ApiKeyError extends Error {  // carries a `code` field
  constructor(code, message) { ... }
}

listAPIKeys()          // GET    → APIKeyResponse[]
createAPIKey(name)     // POST   → {id, name, key, key_prefix, created_at}
revokeAPIKey(id)       // DELETE → void
```

Error codes: `duplicate_name` (409), `name_required` (400), `not_found` (404), `request_failed` (any other non-2xx or network failure). The component maps codes to user-facing copy.

This split exists so the logic is testable: `web/` has vitest but no jsdom or testing-library, and the two existing suites (`src/utils/sw.test.js`, `src/utils/webPush.test.js`) are plain util tests. Keeping `fetch` out of the component preserves that and avoids adding `@testing-library/react` + `jsdom` for a single page.

## 3. Page component — `web/src/components/admin/AdminApiKeys.jsx`

Structure mirrors `AdminUsers.jsx`: `admin-header` with a back-link to `/admin`, a `Key` lucide icon and title, an `admin-actions` bar with a **New API Key** button, and a table.

- **Table** — Name / Key Prefix (monospace) / Created / actions. Empty state: "No API keys yet". Loading state: "Loading API keys…".
- **Create** — the existing `form-overlay` + `form-card` pattern with one required `name` field. Submit disables the button while in flight.
- **Reveal modal** — on `201`, the create form is replaced by a blocking modal: warning that this is the only time the key is shown, the full key in a monospace read-only field, a **Copy** button, the key's name, and a single **I've saved it — close** button. No backdrop-click dismissal, no Esc handler. Copy uses `navigator.clipboard.writeText`, falling back to a hidden textarea plus `document.execCommand('copy')` for non-secure contexts (the portal is served over plain HTTP in local development). Closing the modal clears the plaintext from component state and refetches the list.
- **Revoke** — `window.confirm("Revoke <name>? Any client using it will immediately start getting 401s.")`, then `DELETE`, then refetch. Same shape as `handleDelete` in `AdminUsers.jsx`.
- **Feedback** — `useNotification()`'s `showSuccess`/`showError`, matching the other admin pages.

## 4. Wiring

- `web/src/App.jsx` — add `<Route path="/admin/api-keys" element={<AdminApiKeys />} />` to the admin block.
- `web/src/components/admin/AdminDashboard.jsx` — add an "API Keys" quick action (`Key` icon, `teal`) pointing at the new route.
- `web/src/components/admin/AdminDashboard.css` — define `.action-card.teal` and its `.action-icon` rule. Also define the missing `.action-card.purple` pair: the existing User Access card already passes `color: 'purple'`, but only `.stat-card.purple` was ever written, so that card renders unstyled today. Fixing it is a one-rule change in a file this work already touches.

## 5. Testing

`web/src/utils/apiKeys.test.js`, written before the module, with `global.fetch` stubbed via `vi.fn()`:

- `listAPIKeys` returns the parsed array and requests `/v1/admin/api-keys`.
- `createAPIKey` posts `{name}` with a JSON content-type and returns the parsed body including `key`.
- `createAPIKey` throws `ApiKeyError` with code `duplicate_name` on 409 and `name_required` on 400.
- `revokeAPIKey` issues `DELETE /v1/admin/api-keys/{id}` and throws code `not_found` on 404.
- A network rejection surfaces as code `request_failed`.

These run under the existing `just unit`, which already invokes `pnpm test` in `web/` via the `_web-test` recipe.

Manual verification: `just build`, sign in as an admin, create a key, confirm the modal shows the plaintext and Copy works, reload to confirm the key is listed by prefix only, then revoke it.

## Deliberately excluded

- **`created_by` and `last_used_at` columns.** Neither is in `APIKeyResponse`; `LastUsedAt` was explicitly excluded from the domain model in the auth design. Surfacing either means a backend change and is out of scope here.
- **Editing a key's name.** No `PUT` endpoint exists; revoke-and-recreate is the intended flow for a key.
- **Component-level DOM tests.** Would require `@testing-library/react` and `jsdom`; the utils split covers the logic that can actually break.
