# Embed zensor-ui as a static SPA in zensor-server

## Context

`zensor-ui` (`../zensor-ui`) is a separate React/Vite/Express project that today runs as its
own Node process, proxying `/api/*` to zensor-server's `/v1/*` API and serving a client bundle
(with an unused SSR path). We want to bring the UI into `zensor-server` under `./web`, build it
with Vite, and serve it directly from the Go binary via `go:embed` — one process, one binary,
no separate frontend deployment.

`~/repos/nebula/control-plane` already does this: a `web/` Vite+TS app, built to `dist/`, embedded
by `internal/business/web/embed.go` (`//go:embed all:dist`), served via a `SPAHandler()` registered
at `router.Handle("/", ...)`, with a justfile recipe that hashes `web/` sources to skip redundant
rebuilds. This design follows that pattern, adapted to zensor-ui's existing JS/JSX app instead of
rewriting it in TypeScript.

## Decisions

- **Static SPA only.** Drop zensor-ui's Node/Express layer entirely: no SSR entry
  (`entry-server.jsx`, `server/index.js`), no Node OpenTelemetry instrumentation
  (`server/tracing.js`), no Express API proxy (`server/simple-server.js`), no mock servers
  (`mock-server.js`, `custom-mock-server.js`). The Go server already provides tracing/metrics
  and now serves the UI and API from the same origin, so none of this is needed.
- **Same-origin API calls.** The frontend's `apiBaseUrl` changes from `/api` (previously rewritten
  to `/v1` by the Express proxy) to `/v1` directly. The WebSocket hooks already connect to
  `same-origin` host, and `/ws/device-messages` already exists on the Go server unprefixed, so no
  change needed there.
- **New `/v1/user` endpoint.** zensor-ui's old `/api/user` returned mock data or read
  `Remote-User/Remote-Name/Remote-Email` headers (an oauth2-proxy convention) including a
  `role`/`isAdmin` flag. The Go server has no role concept — it only extracts
  `X-User-ID/X-User-Name/X-User-Email` for tracing (`createUserHeaderMiddleware` in
  `internal/infra/httpserver/server.go`). The new endpoint returns those three fields (empty
  strings if headers are absent) and nothing else.
- **Admin gating removed, not faked.** `useAdmin.js` stops treating "isAdmin" as a real permission
  — the Admin nav link always renders. Nothing enforces it server-side today either way, so this
  doesn't change actual access, just stops implying a check that doesn't exist.
- **Package manager: pnpm.** `./web` uses pnpm (fresh `pnpm-lock.yaml`), not npm.
- **Build integration mirrors nebula.** `justfile` gets `_compute-web-build-hash` / `_web-build`
  recipes that hash `web/`'s tracked sources (+ relevant `VITE_*` env) and skip the pnpm build when
  unchanged. `just build`, `just run`, and `just dev` all depend on `_web-build` before compiling Go.

## Directory layout

```
zensor-server/
  web/                                    ← ported from zensor-ui, pnpm + Vite, static SPA
    src/
      components/...                      (unchanged from zensor-ui, minus SSR-only bits)
      hooks/useAdmin.js                   (updated: no isAdmin gating)
      hooks/useWebSocket.js               (unchanged — already same-origin/native WebSocket)
      config/api.js                       (apiBaseUrl → '/v1')
      main.jsx                            (createRoot entry; entry-client/entry-server.jsx deleted)
      App.jsx, App.css, index.css
    public/
    index.html                           (script src → /src/main.jsx, ssr-outlet comment removed)
    package.json, pnpm-lock.yaml, vite.config.js, eslint.config.js
  internal/infra/httpserver/web/
    embed.go                              ← //go:embed all:dist + SPAHandler (ported from nebula)
  internal/control_plane/httpapi/
    user_controller.go                    ← new: GET /v1/user
```

`internal/infra/httpserver/web/dist/` is build output — gitignored, populated by `just _web-build`.

## Backend wiring

- `embed.go`: near-verbatim copy of nebula's `internal/business/web/embed.go` — `fs.Sub` over the
  embedded `dist`, `http.FileServer`, SPA-fallback (non-asset unknown paths → `/`), asset
  content-type correction for hashed Vite `/assets/*` files, long-cache headers for assets and
  `no-cache` for everything else.
- `internal/infra/httpserver/server.go`'s `NewServer`: add `router.Handle("/", web.SPAHandler())`.
  Go 1.22 `ServeMux` dispatches by pattern specificity, not registration order, so this coexists
  with the existing `/v1/...`, `/ws/...`, `/healthz`, `/metrics` routes without conflict.
- New `user_controller.go` in `internal/control_plane/httpapi/`, following the existing
  `Controller` interface (`AddRoutes(*http.ServeMux)`) convention used by the other controllers in
  that package. `GET /v1/user` reads `X-User-ID`/`X-User-Name`/`X-User-Email` off the request and
  returns them as JSON, using the same response-envelope helper (`ReplyJSONResponse`) the other
  controllers use.
- Wired into Wire DI (`cmd/api/wire/`) alongside the other controllers; `just wire` regenerates
  `wire_gen.go`.

## Frontend changes (relative to zensor-ui)

- **Deleted**: `server/` (whole directory), `mock-server.js`, `custom-mock-server.js`,
  `src/entry-server.jsx`, `src/entry-client.jsx`, `.dockerignore`/`Dockerfile` (frontend no longer
  ships as its own container), `custom-mock-server` / OTel / SSR npm scripts and dependencies.
- **Kept as-is**: all `src/components/`, `src/hooks/useWebSocket.js`, `src/hooks/useNotification.js`,
  CSS, `public/vite.svg`.
- **Changed**:
  - `src/main.jsx` becomes the sole entry point (`createRoot`, no `StrictMode` change needed —
    already there); `index.html` points to it and drops the SSR outlet comment.
  - `src/config/api.js`: `apiBaseUrl` default `/api` → `/v1`.
  - `src/hooks/useAdmin.js` and `src/components/UserInfo.jsx`: fetch `/v1/user` instead of
    `/api/user`; `useAdmin.js` no longer derives `isAdmin` from the response (see Decisions).
  - `vite.config.js`: single `index.html` entry, `outDir: 'dist'` (was split
    `dist/client`/`dist/server`), `ssr` config block removed, `server.proxy` added for `/v1` and
    `/ws` → `http://localhost:3000` for local Vite-dev-server usage.
  - `package.json`: pnpm; dependencies trimmed to `react`, `react-dom`, `react-router-dom`,
    `lucide-react`; devDependencies trimmed to `vite`, `@vitejs/plugin-react`, eslint + plugins.
    Scripts reduced to `dev`, `build`, `lint`, `preview`.

## Build tooling (justfile)

Mirrors nebula's pattern:

- `_compute-web-build-hash`: hashes `git ls-files -co --exclude-standard -- web/` plus relevant
  `VITE_*` env vars.
- `_web-build`: compares against a stamp file in `internal/infra/httpserver/web/dist/`, runs
  `pnpm -C web install --frozen-lockfile && pnpm -C web run build` only when the hash changed,
  writes the new hash to the stamp on success.
- `just build`, `just run`, and `just dev` (the recently-added seed-data dev command) all call
  `just _web-build` before building/running the Go binary.

## Error handling

- `SPAHandler`: non-asset unknown paths (`/tenants/123/devices`, etc.) fall back to `index.html` so
  client-side routing handles them; unknown `/assets/*` paths 404 for real.
- `/v1/user`: never errors — missing headers just yield empty-string fields. No upstream
  auth proxy is configured in this deployment today, so this is the honest default, not a
  regression.

## Testing

- Go: Ginkgo unit test for the new `/v1/user` handler (headers present / absent), following
  existing `httpapi` controller test conventions (`test/unit/doubles/...` mocks not needed — this
  handler has no dependencies beyond the request).
- Frontend: zensor-ui has no existing JS test suite (no vitest/testing-library in its
  `package.json`) — none added here; out of scope for a port.
- Manual verification: `just build` compiles with the embed; smoke-test the SPA served from the Go
  binary — tenant list, a device page, `/v1/user`, and the live-messages WebSocket page — against
  the real backend.

## Out of scope

- Any visual/UX redesign of the ported components.
- Adding a role/permission system to zensor-server (the admin-gating removal is a scope
  reduction, not a new feature).
- CI/CD changes for building/publishing this UI as a separate artifact — it now ships inside the
  existing Go build/deploy pipeline.
