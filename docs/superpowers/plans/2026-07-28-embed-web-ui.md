# Embed zensor-ui as a static SPA — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port `../zensor-ui` into `./web`, build it with Vite/pnpm as a static SPA, and embed it into the zensor-server Go binary via `go:embed`, following `~/repos/nebula/control-plane`'s embed pattern.

**Architecture:** A trimmed, CSR-only copy of zensor-ui lives in `./web` (pnpm + Vite, no Express/SSR/Node-OTel). Vite builds directly into `internal/infra/httpserver/web/dist`, which a new `web` Go package embeds via `//go:embed all:dist` and serves through a `SPAHandler()` registered at `router.Handle("/", ...)` in the existing `NewServer`. A new `GET /v1/me` endpoint replaces the old mock `/api/user`, and the frontend talks to the API same-origin under `/v1/*` instead of through the old `/api` → `/v1` Express proxy.

**Tech Stack:** Go 1.24 (existing), React 19 + react-router-dom v7 + Vite 6 (existing zensor-ui deps, trimmed), pnpm (already installed at v10.33.0), Ginkgo v2/Gomega for Go tests.

**Spec:** `docs/superpowers/specs/2026-07-28-embed-web-ui-design.md`

## Global Constraints

- Package manager for `./web` is **pnpm** — never npm/yarn.
- The "current user" endpoint is `GET /v1/me` (not `/v1/user` — that path would be confusable with the existing `GET /v1/users/{id}` route in `internal/shared_kernel/httpapi/user_controller.go`), returning `{"user_id","name","email"}` sourced from the `X-User-ID`/`X-User-Name`/`X-User-Email` headers. No `isAdmin`/role field — admin gating is removed, not faked (per approved design).
- Vite's `build.outDir` for `./web` is `../internal/infra/httpserver/web/dist` (relative to `web/`), with `emptyOutDir: true`. This is the only place the SPA is built to — there is no separate `web/dist`.
- `internal/infra/httpserver/web/dist/` is never committed — it's gitignored and produced by `just _web-build`. Any bare `go build`/`go test` on the `internal/infra/httpserver/web` package (or anything importing it) will fail to compile until that directory exists with real content — this is expected and matches nebula's setup. Always go through `just build`/`just unit`/`just run`/`just dev`, which all depend on `_web-build`.
- Follow existing repo conventions: `slog` global logger, `context.Context` first param for I/O, Ginkgo v2 BDD tests, `ReplyJSONResponse`/`Controller` helpers in `internal/infra/httpserver`, comments only on exported items.
- Go 1.22+ `http.ServeMux` dispatches by pattern specificity, not registration order — registering `router.Handle("/", ...)` alongside existing `/v1/...` and `/ws/...` patterns is safe regardless of where it's added in `NewServer`.

---

### Task 1: Add `GET /v1/me` endpoint to the Go server

**Files:**
- Modify: `internal/infra/httpserver/server.go`
- Modify: `internal/infra/httpserver/server_test.go`

**Interfaces:**
- Produces: `CurrentUserResponse{UserID, Name, Email string}` (JSON tags `user_id`, `name`, `email`), route `GET /v1/me`. Later tasks (frontend) consume this as the JSON body of a same-origin `fetch('/v1/me')`.

- [ ] **Step 1: Write the failing tests**

Add to `internal/infra/httpserver/server_test.go` (existing imports already include `net/http`, `net/http/httptest`; add `encoding/json`):

```go
import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)
```

Add this `Context` inside the top-level `Describe("HTTPServer", ...)` block, alongside the existing `Context`s:

```go
	ginkgo.Context("CurrentUser", func() {
		ginkgo.When("the request carries user headers", func() {
			ginkgo.It("should return them as JSON", func() {
				req := httptest.NewRequest("GET", "/v1/me", nil)
				req.Header.Set("X-User-ID", "user123")
				req.Header.Set("X-User-Name", "John Doe")
				req.Header.Set("X-User-Email", "john.doe@example.com")
				rec := httptest.NewRecorder()

				getCurrentUser().ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

				var body CurrentUserResponse
				gomega.Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body).To(gomega.Equal(CurrentUserResponse{
					UserID: "user123",
					Name:   "John Doe",
					Email:  "john.doe@example.com",
				}))
			})
		})

		ginkgo.When("the request has no user headers", func() {
			ginkgo.It("should return empty fields, not an error", func() {
				req := httptest.NewRequest("GET", "/v1/me", nil)
				rec := httptest.NewRecorder()

				getCurrentUser().ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))

				var body CurrentUserResponse
				gomega.Expect(json.Unmarshal(rec.Body.Bytes(), &body)).To(gomega.Succeed())
				gomega.Expect(body).To(gomega.Equal(CurrentUserResponse{}))
			})
		})
	})
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `just unit internal/infra/httpserver`
Expected: FAIL to compile — `undefined: getCurrentUser`, `undefined: CurrentUserResponse`.

- [ ] **Step 3: Implement the handler**

In `internal/infra/httpserver/server.go`, add the route registration right after the existing healthz/metrics routes inside `NewServer`:

```go
	router.Handle("GET /healthz", getHealthz())
	router.Handle("GET /metrics", promhttp.Handler())
	router.Handle("GET /v1/me", getCurrentUser())
```

Add the handler and response type near `getHealthz`/`HealthzResponse` at the bottom of the file:

```go
func getCurrentUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		output := CurrentUserResponse{
			UserID: r.Header.Get("X-User-ID"),
			Name:   r.Header.Get("X-User-Name"),
			Email:  r.Header.Get("X-User-Email"),
		}
		ReplyJSONResponse(w, http.StatusOK, output)
	}
}

type CurrentUserResponse struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `just unit internal/infra/httpserver`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/infra/httpserver/server.go internal/infra/httpserver/server_test.go
git commit -m "feat: add GET /v1/me endpoint returning current user headers"
```

---

### Task 2: Scaffold the `./web` frontend app

**Files:**
- Create: `web/src/**` (ported from `/Users/sdiaz/repos/zensor/zensor-ui/src`, minus SSR entries)
- Create: `web/public/vite.svg`
- Create: `web/package.json`, `web/vite.config.js`, `web/index.html`, `web/eslint.config.js`, `web/.gitignore`
- Modify: `web/src/main.jsx` (add missing `BrowserRouter`)
- Modify: `/Users/sdiaz/repos/zensor/zensor-server/.gitignore` (add `web/node_modules/`)

**Interfaces:**
- Produces: a standalone, buildable Vite React app at `./web` that builds to `../internal/infra/httpserver/web/dist`. Task 4 embeds that directory. Task 3 edits this app's API calls in place.

- [ ] **Step 1: Copy source files from zensor-ui, excluding SSR-only entries**

```bash
cd /Users/sdiaz/repos/zensor/zensor-server
mkdir -p web
cp -r /Users/sdiaz/repos/zensor/zensor-ui/src web/src
cp -r /Users/sdiaz/repos/zensor/zensor-ui/public web/public
cp /Users/sdiaz/repos/zensor/zensor-ui/eslint.config.js web/eslint.config.js
rm web/src/entry-client.jsx web/src/entry-server.jsx
```

This brings over `web/src/{App.jsx,App.css,index.css,main.jsx,assets/,components/,config/,hooks/}` and `web/public/vite.svg` unchanged.

- [ ] **Step 2: Write the trimmed `web/package.json`**

```json
{
  "name": "zensor-server-web",
  "private": true,
  "version": "0.0.0",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "lint": "eslint . --ext js,jsx --report-unused-disable-directives --max-warnings 0",
    "preview": "vite preview"
  },
  "dependencies": {
    "lucide-react": "^0.513.0",
    "react": "^19.1.0",
    "react-dom": "^19.1.0",
    "react-router-dom": "^7.6.2"
  },
  "devDependencies": {
    "@eslint/js": "^9.25.0",
    "@vitejs/plugin-react": "^4.4.1",
    "eslint": "^9.25.0",
    "eslint-plugin-react-hooks": "^5.2.0",
    "eslint-plugin-react-refresh": "^0.4.19",
    "globals": "^16.0.0",
    "vite": "^6.3.5"
  }
}
```

This drops zensor-ui's `express`, `cors`, `dotenv`, `http-proxy-middleware`, `pino`/`pino-pretty`, all `@opentelemetry/*`, `uuid`, and `ws` — none of these are imported anywhere under `src/` (verified: `uuid`/`ws` packages aren't referenced; only the browser-global `WebSocket` is used in `src/hooks/useWebSocket.js` and `src/components/admin/AdminHealth.jsx`).

- [ ] **Step 3: Write `web/vite.config.js`**

```js
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: '../internal/infra/httpserver/web/dist',
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/v1': 'http://localhost:3000',
      '/ws': {
        target: 'ws://localhost:3000',
        ws: true,
      },
    },
  },
})
```

The `server.proxy` block is what makes `pnpm -C web run dev` usable against a real, already-running Go server (`just run` / `just dev`) for local frontend development with hot reload.

- [ ] **Step 4: Write `web/index.html`**

```html
<!doctype html>
<html lang="en">

<head>
  <meta charset="UTF-8" />
  <link rel="icon" type="image/svg+xml" href="/vite.svg" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <meta http-equiv="Content-Security-Policy"
    content="default-src 'self' http: https: data: blob: 'unsafe-inline'; connect-src 'self' http: https: wss: ws: data: blob:;">
  <title>Zensor Portal</title>
</head>

<body>
  <div id="root"></div>
  <script type="module" src="/src/main.jsx"></script>
</body>

</html>
```

(Same as zensor-ui's `index.html`, minus the `<!--ssr-outlet-->` comment inside `#root`, since there's no SSR.)

- [ ] **Step 5: Fix `web/src/main.jsx` — it's missing the router**

zensor-ui's `index.html` actually pointed at `src/entry-client.jsx` (which wrapped `<App />` in `<BrowserRouter>` via `hydrateRoot`, for the SSR path). `src/main.jsx` was a vestigial create-vite-template file that was never wired up and does **not** wrap `<App />` in a router — but `App.jsx` uses `useLocation()` and `<Routes>`, which throw without one. Now that `main.jsx` is the sole entry, it must provide the router:

Read the current `web/src/main.jsx` first, then replace its contents:

```jsx
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import './index.css'
import App from './App.jsx'

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </StrictMode>,
)
```

- [ ] **Step 6: Write `web/.gitignore`**

```
node_modules/
.vite/
```

- [ ] **Step 7: Add `web/node_modules/` and the embed output dir to the root `.gitignore`**

Read `/Users/sdiaz/repos/zensor/zensor-server/.gitignore` first, then append:

```

# Frontend (web/)
web/node_modules/

# Vite build output for //go:embed — run `just _web-build` (or `just build`) before go build/test.
internal/infra/httpserver/web/dist/
```

(This is added now, before Step 8 first creates `internal/infra/httpserver/web/dist/`, so it never shows up as an untracked build artifact in `git status`.)

- [ ] **Step 8: Install dependencies and verify the app builds standalone**

```bash
cd /Users/sdiaz/repos/zensor/zensor-server
pnpm -C web install
pnpm -C web run build
```

Expected: build succeeds, producing `internal/infra/httpserver/web/dist/index.html` and `internal/infra/httpserver/web/dist/assets/*`. (The Go side doesn't reference this directory yet — that's Task 4 — so this just proves the trimmed app compiles.)

- [ ] **Step 9: Commit**

```bash
git add web .gitignore
git commit -m "feat: scaffold ./web — pnpm/Vite port of zensor-ui, CSR-only"
```

---

### Task 3: Repoint frontend API calls to same-origin `/v1`, remove admin-role fakery

**Files:**
- Modify: `web/src/config/api.js`
- Modify: `web/src/components/UserInfo.jsx`
- Modify: `web/src/components/Profile.jsx`
- Modify: `web/src/App.jsx`
- Delete: `web/src/hooks/useAdmin.js`
- Modify: `web/src/components/admin/AdminCommands.jsx`, `AdminDashboard.jsx`, `AdminDevices.jsx`, `AdminHealth.jsx`, `AdminScheduledTasks.jsx`, `AdminTaskExecutions.jsx`, `AdminTenants.jsx`

**Interfaces:**
- Consumes: `GET /v1/me` from Task 1 (`{"user_id","name","email"}`).
- Produces: no new interfaces — this task only changes which URLs the existing UI code calls.

- [ ] **Step 1: Change the default API base URL**

In `web/src/config/api.js`, change:

```js
    apiBaseUrl: import.meta.env.VITE_API_BASE_URL || '/api',
```

to:

```js
    apiBaseUrl: import.meta.env.VITE_API_BASE_URL || '/v1',
```

This covers every call site that goes through `getApiUrl`/`buildApiEndpoint` (the `TenantList`, `TenantDevices`, `TenantDeviceCard`, `TenantPortal`, `TenantListWithNotifications` components, and the `scheduledTasksApi`/`deviceCommandsApi` helpers in this same file).

- [ ] **Step 2: Point `UserInfo.jsx` at `/v1/me`**

In `web/src/components/UserInfo.jsx`, change:

```js
                const response = await fetch('/api/user')
```

to:

```js
                const response = await fetch('/v1/me')
```

- [ ] **Step 3: Point `Profile.jsx` at `/v1/me` and drop the fake admin role**

Read `web/src/components/Profile.jsx` first. Make these changes:

a) Change the fetch call:

```js
                const response = await fetch('/api/user')
```
→
```js
                const response = await fetch('/v1/me')
```

b) Remove the derived role line:

```js
    const role = userInfo.isAdmin ? 'Admin' : 'User'
```
→ delete this line entirely.

c) Remove the role subtitle under the user's name:

```jsx
                    <div className="profile-title">
                        <h3>{userInfo.name || 'Unknown Name'}</h3>
                        <p className="profile-subtitle">{role}</p>
                    </div>
```
→
```jsx
                    <div className="profile-title">
                        <h3>{userInfo.name || 'Unknown Name'}</h3>
                    </div>
```

d) Remove the "Role" field block entirely:

```jsx
                    <div className="profile-field">
                        <div className="field-label">
                            <Shield size={16} />
                            Role
                        </div>
                        <div className="field-value">
                            <span className={`role-badge ${userInfo.isAdmin ? 'admin' : 'user'}`}>
                                {role}
                            </span>
                        </div>
                    </div>
```
→ delete this block entirely (the `Shield` import stays — it's still used for the "Tenant" field label a few lines below).

e) Replace every remaining literal `/api/` prefix in this file with `/v1/` (these are the `/api/tenants/{id}/configuration` and `/api/users/{email}` calls, unrelated to the user-role removal):

```bash
cd /Users/sdiaz/repos/zensor/zensor-server/web/src/components
sed -i '' "s#/api/#/v1/#g" Profile.jsx
```

- [ ] **Step 4: Remove the admin nav gating from `App.jsx`**

Read `web/src/App.jsx` first. Remove the import:

```js
import { useAdmin } from './hooks/useAdmin'
```

Remove this line inside the `App` function:

```js
  const { isAdmin, isLoading } = useAdmin()
```

Change the conditionally-rendered Admin nav link:

```jsx
                  {!isLoading && isAdmin && (
                    <Link to="/admin" className="nav-link admin-link">
                      <Shield size={20} />
                      Admin
                    </Link>
                  )}
```

to an unconditional one:

```jsx
                  <Link to="/admin" className="nav-link admin-link">
                    <Shield size={20} />
                    Admin
                  </Link>
```

Nothing else changes — the `/admin/*` routes were already unconditionally registered in the `<Routes>` block below (the nav link was only ever a cosmetic gate, never an access control mechanism; direct navigation to `/admin` already worked before this change).

- [ ] **Step 5: Remove `useAdmin` gating from the seven admin page components, then delete the hook**

Discovered during implementation: it's not just `App.jsx` that uses `useAdmin()` — all seven files in `web/src/components/admin/` (`AdminDashboard.jsx`, `AdminTenants.jsx`, `AdminDevices.jsx`, `AdminCommands.jsx`, `AdminScheduledTasks.jsx`, `AdminTaskExecutions.jsx`, `AdminHealth.jsx`) import it too, each with the same four-part pattern:

1. `import { useAdmin } from '../../hooks/useAdmin'`
2. `const { isAdmin, isLoading } = useAdmin()`
3. Inside a `useEffect`: an early-return block `if (!isLoading && !isAdmin) { showError(...); return }`, followed by the real fetch guarded by `if (isAdmin [&& other conditions like tenantId/deviceId]) { fetchX() }`, with `isAdmin`/`isLoading` in the dependency array
4. Two render-time guards further down: `if (isLoading) { return <...loading UI...> }` then `if (!isAdmin) { return <...Access Denied UI...> }`

Leaving these as-is would make every Admin page permanently broken (not neutral) once `useAdmin` no longer receives a real `isAdmin` value — every page would show "Access denied" forever and never fetch data. This isn't consistent with the approved design decision ("useAdmin.js drops the isAdmin concept" — removed everywhere, not faked anywhere), so it must be removed from all seven files, not just `App.jsx`.

For each of the seven files, apply the same four-part removal:
1. Delete the `useAdmin` import line.
2. Delete the `const { isAdmin, isLoading } = useAdmin()` line.
3. In the `useEffect`: delete the `if (!isLoading && !isAdmin) { ...; return }` block entirely. Change the `if (isAdmin && <other conditions>) { fetchX() }` line to keep only the `<other conditions>` (e.g. `if (tenantId) { fetchX() }`, or call `fetchX()` unconditionally if there were no other conditions). Remove `isAdmin`/`isLoading` from that `useEffect`'s dependency array (keep any other dependencies, e.g. `tenantId`, `deviceId`, `taskId`).
4. Delete both render-time guard blocks: the `if (isLoading) { ... }` block and the `if (!isAdmin) { ... }` block, in that order, so the component falls straight through to its normal render.

Then delete the now-fully-unused hook:

```bash
rm /Users/sdiaz/repos/zensor/zensor-server/web/src/hooks/useAdmin.js
```

Verify nothing still imports it: `grep -rn useAdmin web/src` must print nothing.

- [ ] **Step 6: Bulk-rename remaining hardcoded `/api/` calls in the admin components**

These seven files call `fetch('/api/...')` directly (not through `config/api.js`), for tenants/devices/tasks/scheduled-tasks endpoints only — none of them call `/api/user`, so a straight prefix rename is safe:

```bash
cd /Users/sdiaz/repos/zensor/zensor-server/web/src/components/admin
sed -i '' "s#/api/#/v1/#g" AdminCommands.jsx AdminDashboard.jsx AdminDevices.jsx AdminHealth.jsx AdminScheduledTasks.jsx AdminTaskExecutions.jsx AdminTenants.jsx
```

- [ ] **Step 7: Verify nothing was missed and the app still builds/lints**

```bash
cd /Users/sdiaz/repos/zensor/zensor-server
grep -rn "'/api/\|\`/api/" web/src || echo "no remaining /api/ references"
pnpm -C web run lint
pnpm -C web run build
```

Expected: the grep prints nothing but the "no remaining" message; lint and build both succeed.

- [ ] **Step 8: Commit**

```bash
git add web
git commit -m "refactor: point web UI at same-origin /v1 API, drop fake admin gating"
```

---

### Task 4: Embed the built SPA into the Go server

**Files:**
- Create: `web/public/assets/test-fixture.css` (test fixture only, copied verbatim into `dist/assets/` by Vite)
- Create: `internal/infra/httpserver/web/embed.go`
- Create: `internal/infra/httpserver/web/embed_test.go`
- Create: `internal/infra/httpserver/web/suite_test.go`
- Modify: `internal/infra/httpserver/server.go`
- Modify: `internal/infra/httpserver/server_test.go`

**Interfaces:**
- Consumes: the Vite build output at `internal/infra/httpserver/web/dist` (produced by `pnpm -C web run build`, per Task 2/3's `vite.config.js`).
- Produces: `web.SPAHandler() http.Handler` (exported), registered at `router.Handle("/", web.SPAHandler())` in `NewServer`. No later task consumes this directly — it's the terminal integration point.

- [ ] **Step 1: Add a test fixture asset and rebuild the frontend**

```bash
mkdir -p /Users/sdiaz/repos/zensor/zensor-server/web/public/assets
cat > /Users/sdiaz/repos/zensor/zensor-server/web/public/assets/test-fixture.css <<'EOF'
/* fixture asset for internal/infra/httpserver/web/embed_test.go — not part of the app UI */
body { color: red; }
EOF
cd /Users/sdiaz/repos/zensor/zensor-server
pnpm -C web run build
```

Expected: `internal/infra/httpserver/web/dist/assets/test-fixture.css` and `internal/infra/httpserver/web/dist/index.html` now exist.

- [ ] **Step 2: Write the failing embed package tests**

Create `internal/infra/httpserver/web/suite_test.go`:

```go
package web_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
)

func TestWeb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Web Suite")
}
```

Create `internal/infra/httpserver/web/embed_test.go`:

```go
package web_test

import (
	"net/http"
	"net/http/httptest"

	"zensor-server/internal/infra/httpserver/web"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

var _ = Describe("SPAHandler", func() {
	var recorder *httptest.ResponseRecorder

	BeforeEach(func() {
		recorder = httptest.NewRecorder()
	})

	Context("known asset", func() {
		When("the asset exists in dist/assets", func() {
			It("should serve it with the correct content type", func() {
				request := httptest.NewRequest(http.MethodGet, "/assets/test-fixture.css", nil)

				web.SPAHandler().ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(recorder.Header().Get("Content-Type")).To(gomega.HavePrefix("text/css"))
			})
		})

		When("the asset does not exist", func() {
			It("should return 404, not the SPA fallback", func() {
				request := httptest.NewRequest(http.MethodGet, "/assets/missing.js", nil)

				web.SPAHandler().ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusNotFound))
				gomega.Expect(recorder.Body.String()).NotTo(gomega.ContainSubstring("<!DOCTYPE html>"))
			})
		})
	})

	Context("client-side route", func() {
		When("the path is not a known asset or file", func() {
			It("should fall back to index.html", func() {
				request := httptest.NewRequest(http.MethodGet, "/tenants/example/devices", nil)

				web.SPAHandler().ServeHTTP(recorder, request)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(recorder.Body.String()).To(gomega.ContainSubstring("<!DOCTYPE html>"))
			})
		})
	})
})
```

- [ ] **Step 3: Run the tests to verify they fail**

Run: `go test ./internal/infra/httpserver/web/...`
Expected: FAIL to compile — `no required module provides package .../internal/infra/httpserver/web` / `undefined: web.SPAHandler` (the `web` package doesn't exist yet).

- [ ] **Step 4: Implement `embed.go`**

Create `internal/infra/httpserver/web/embed.go`:

```go
package web

import (
	"embed"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// SPAHandler serves the Vite-built React SPA embedded in dist. Unknown non-asset paths fall
// back to index.html so client-side routing (react-router) handles them; unknown /assets/*
// paths 404 for real.
func SPAHandler() http.Handler {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlPath := r.URL.Path

		isAsset := strings.HasPrefix(urlPath, "/assets/")
		if isAsset {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			if ct := assetContentType(urlPath); ct != "" {
				w = &contentTypeResponseWriter{ResponseWriter: w, contentType: ct}
			}
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}

		if urlPath != "/" {
			fsPath := strings.TrimPrefix(urlPath, "/")
			if _, err := fs.Stat(sub, fsPath); err != nil {
				if isAsset {
					http.NotFound(w, r)
					return
				}
				r.URL.Path = "/"
			}
		}

		fileServer.ServeHTTP(w, r)
	})
}

func assetContentType(urlPath string) string {
	if !strings.HasPrefix(urlPath, "/assets/") {
		return ""
	}
	return mime.TypeByExtension(path.Ext(urlPath))
}

type contentTypeResponseWriter struct {
	http.ResponseWriter
	contentType string
}

func (w *contentTypeResponseWriter) WriteHeader(statusCode int) {
	if w.contentType != "" {
		w.ResponseWriter.Header().Set("Content-Type", w.contentType)
	}
	w.ResponseWriter.WriteHeader(statusCode)
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/infra/httpserver/web/...`
Expected: PASS.

- [ ] **Step 6: Wire the SPA handler into `NewServer`, with a failing-then-passing integration test**

First, add this `Context` to `internal/infra/httpserver/server_test.go` (inside the top-level `Describe("HTTPServer", ...)`), and add the import `"zensor-server/internal/infra/httpserver/web"`:

```go
	ginkgo.Context("StaticWebUI", func() {
		ginkgo.When("requesting the root path", func() {
			ginkgo.It("should serve the embedded SPA", func() {
				srv := NewServer()
				req := httptest.NewRequest("GET", "/", nil)
				rec := httptest.NewRecorder()

				srv.server.Handler.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(rec.Body.String()).To(gomega.ContainSubstring("<!DOCTYPE html>"))
			})
		})

		ginkgo.When("requesting a real API route with the SPA also registered", func() {
			ginkgo.It("should still route to healthz, not the SPA fallback", func() {
				srv := NewServer()
				req := httptest.NewRequest("GET", "/healthz", nil)
				rec := httptest.NewRecorder()

				srv.server.Handler.ServeHTTP(rec, req)

				gomega.Expect(rec.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(rec.Header().Get("Content-Type")).To(gomega.Equal("application/json"))
			})
		})
	})
```

Run: `just unit internal/infra/httpserver`
Expected: FAIL — `/` currently 404s (no handler registered for it yet).

Now add the import and registration to `internal/infra/httpserver/server.go`:

```go
import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"zensor-server/internal/infra/httpserver/web"
	"zensor-server/internal/infra/node"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	_ "net/http/pprof"
)
```

And in `NewServer`, register it (order doesn't matter for `http.ServeMux`, but keep it near the other top-level routes for readability):

```go
	router.Handle("GET /healthz", getHealthz())
	router.Handle("GET /metrics", promhttp.Handler())
	router.Handle("GET /v1/me", getCurrentUser())
	router.Handle("/", web.SPAHandler())
```

- [ ] **Step 7: Run the tests to verify they pass**

Run: `just unit internal/infra/httpserver`
Expected: PASS (both the new `StaticWebUI` tests and every pre-existing test in the package).

- [ ] **Step 8: Commit**

`internal/infra/httpserver/web/dist/` is already gitignored (Task 2, Step 7), so `git add internal/infra/httpserver/web` only stages `embed.go`, `embed_test.go`, and `suite_test.go`.

```bash
git add internal/infra/httpserver/web web/public/assets/test-fixture.css internal/infra/httpserver/server.go internal/infra/httpserver/server_test.go
git commit -m "feat: embed the built web SPA and serve it from the Go server"
```

---

### Task 5: Wire the web build into the justfile

**Files:**
- Modify: `justfile`

**Interfaces:**
- Consumes: `web/` sources (Task 2/3), `internal/infra/httpserver/web/dist` as the build target (Task 4).
- Produces: `just _web-build` recipe, depended on by `build`, `run`, `dev`, and `unit`.

- [ ] **Step 1: Add the hash and build recipes**

Read `justfile` first. Add these two recipes near the top (after `install-otelcol`, before `build`):

```
# Fingerprint of web/ sources (used to skip redundant pnpm builds).
_compute-web-build-hash:
    #!/usr/bin/env bash
    set -euo pipefail
    {
        while IFS= read -r f; do
            shasum -a 256 "$f"
        done < <(git ls-files -co --exclude-standard -- web/ | sort)
    } | shasum -a 256 | awk '{print $1}'

# Vite output for //go:embed (required before any go build or ./internal/... tests).
_web-build:
    #!/usr/bin/env bash
    set -euo pipefail
    hash="$(just _compute-web-build-hash)"
    stamp="internal/infra/httpserver/web/dist/.web-build-hash"
    if [[ -f "$stamp" && "$(cat "$stamp")" == "$hash" && -f internal/infra/httpserver/web/dist/index.html ]]; then
        echo "web build: up to date ($hash)"
        exit 0
    fi
    pnpm -C web install --frozen-lockfile
    out=$(pnpm -C web run build 2>&1)
    code=$?
    if [[ $code -ne 0 ]]; then echo "$out"; exit $code; fi
    mkdir -p internal/infra/httpserver/web/dist
    echo "$hash" > "$stamp"
    duration=$(echo "$out" | grep -oE "built in [0-9ms.]+" | tail -1)
    echo "web build: done ($duration)"
```

- [ ] **Step 2: Depend on `_web-build` from `build`, `run`, `dev`, and `unit`**

Change:

```
build:
    go build -o server cmd/api/main.go
```
to:
```
build:
    just _web-build
    go build -o server cmd/api/main.go
```

`run: build` and `dev: build` already declare `build` as a prerequisite (justfile dependency syntax), and `build` now runs `_web-build` first — so both `just run` and `just dev` pick up the web build transitively. No changes needed to the `run` or `dev` recipe bodies.

Change:

```
unit path="internal":
    go run github.com/onsi/ginkgo/v2/ginkgo run -r --randomize-all --randomize-suites --fail-on-pending --keep-going --cover --coverprofile=coverprofile.out --race --trace --timeout=4m {{path}}
```
to:
```
unit path="internal":
    just _web-build
    go run github.com/onsi/ginkgo/v2/ginkgo run -r --randomize-all --randomize-suites --fail-on-pending --keep-going --cover --coverprofile=coverprofile.out --race --trace --timeout=4m {{path}}
```

(`unit` doesn't depend on `build` today, so it needs its own explicit dependency — otherwise `just unit` fails to compile `internal/infra/httpserver/web` from a clean checkout.)

- [ ] **Step 3: Verify**

```bash
rm -rf internal/infra/httpserver/web/dist web/node_modules
just build
ls internal/infra/httpserver/web/dist/index.html
just build   # second run should hit the cache
```

Expected: first `just build` runs `pnpm install` + `pnpm run build` (prints "web build: done ..."), produces `internal/infra/httpserver/web/dist/index.html`, then compiles the Go binary. The second `just build` prints "web build: up to date ..." and skips the pnpm build.

Then:

```bash
just unit internal/infra/httpserver
```

Expected: PASS (this also exercises the `_web-build` dependency on a path that previously failed to compile without it).

- [ ] **Step 4: Commit**

```bash
git add justfile
git commit -m "build: wire web app build into justfile (build/run/dev/unit)"
```

---

### Task 6: End-to-end verification

**Files:** none (verification only).

- [ ] **Step 1: Full build**

```bash
cd /Users/sdiaz/repos/zensor/zensor-server
just build
```

Expected: succeeds, produces `./server`.

- [ ] **Step 2: Full unit suite**

```bash
just unit
```

Expected: all packages pass, including the new `internal/infra/httpserver` and `internal/infra/httpserver/web` tests.

- [ ] **Step 3: Run the server and smoke-test the embedded UI**

```bash
ENV=local ./server &
SERVER_PID=$!
sleep 1
curl -sf http://127.0.0.1:3000/healthz
curl -sf http://127.0.0.1:3000/v1/me
curl -sf http://127.0.0.1:3000/ | grep -o '<title>[^<]*</title>'
curl -sf http://127.0.0.1:3000/tenants/nonexistent/devices | grep -o '<title>[^<]*</title>'
kill $SERVER_PID
```

Expected:
- `/healthz` returns the existing JSON health payload.
- `/v1/me` returns `{"user_id":"","name":"","email":""}` (no headers set on a plain curl).
- `/` and the client-route path both return the SPA's `index.html` (`<title>Zensor Portal</title>`), proving the SPA fallback works and doesn't shadow real routes.

- [ ] **Step 4: Manual browser smoke test**

Run `just dev` (builds, starts Docker deps unless `ENV=local`, seeds data, tails logs) and open `http://localhost:3000/` in a browser. Verify:
- The tenant list loads (calls `/v1/tenants`).
- Navigating into a tenant's devices page works.
- `/live-messages` opens a WebSocket to `/ws/device-messages` without errors in the browser console.
- The user menu (top right) shows "Guest" (no `X-User-ID` header set locally) rather than erroring.
- `/admin` is reachable directly and its nav link is visible (no more admin gating).

No task changes are needed here regardless of outcome — this step is purely to confirm the port behaves as designed before considering the work done. If something doesn't match, treat it as a bug against the specific task above that owns the broken piece, not a new scope item.
