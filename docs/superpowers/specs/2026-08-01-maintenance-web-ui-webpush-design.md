# Maintenance Module Web UI + Native Browser Push — Design

**Date:** 2026-08-01
**Status:** Approved

## Goal

Bring the zensor-app (Flutter) maintenance features into zensor-server's embedded web UI, and deliver push notifications to browsers via native Web Push (VAPID) while keeping FCM for mobile tokens.

## Context

- The backend already exposes every maintenance endpoint the mobile app uses (`/v1/maintenance/activities*`, `/v1/maintenance/executions*`), plus delete and activate/deactivate that the app never calls.
- The embedded React UI (`web/src/`) has zero maintenance coverage today.
- Push delivery is FCM HTTP v1 (`internal/infra/notification/fcm_client.go`) behind `CompositeNotificationClient`; `PushNotificationWorker` (`internal/maintenance/usecases/push_notification_worker.go`) fans reminder/overdue events out to all tenant users' tokens registered via `POST /v1/users/{id}/push-tokens`.
- `domain.PushToken` already stores `Platform` (`ios`/`android`/`web`).

## Decisions (user-confirmed)

1. Browser push = **native Web Push with VAPID** (no Firebase in the browser).
2. **Route by platform**: FCM stays for `ios`/`android` tokens; `web` tokens go through the new web push client.
3. Maintenance UI lives in the **tenant portal** (user-facing), not admin.
4. Scope = full app parity **plus** activity delete + pause/resume, custom `field_values` captured on completion, and real list pagination.
5. Push enrollment = **toggle on the Profile page**.
6. Server shape = **platform-routing client**; subscription JSON stored as the token string in the existing `push_tokens` table (no schema change).

## 1. Backend — web push channel

### WebPushClient
- New `internal/infra/notification/webpush_client.go` using `github.com/SherClockHolmes/webpush-go`.
- The browser's `PushSubscription` JSON (`{endpoint, keys: {p256dh, auth}}`) is stored verbatim as `PushToken.Token` with `platform: "web"`.
- `SendPushNotification` unmarshals the subscription from `request.Token` and sends a JSON payload `{"title", "body", "deeplink"}`.
- `SendEmail` returns an error by design (mirrors `FCMClient`).
- A 404/410 response from the push service returns sentinel `ErrSubscriptionExpired`.

### Platform routing
- `notification.PushNotificationRequest` gains `Platform string`.
- New `PlatformRoutingPushClient{fcm, webpush NotificationClient}`: `Platform == "web"` → webpush client; anything else (ios/android/empty) → FCM. Plugged in as the composite's push client via Wire.
- `PushNotificationWorker` and `user_push_broadcast_sender.go` populate `Platform` from each `domain.PushToken`.
- On `ErrSubscriptionExpired`, the caller unregisters the token via `PushTokenService.UnregisterToken` so dead subscriptions self-clean.

### Config & endpoint
- `config/server.yaml`: `notification.webpush.{vapid_public_key, vapid_private_key, subscriber}` (env-overridable via `ZENSOR_SERVER_*`).
- New endpoint `GET /v1/push/vapid-public-key` → `{"public_key": "..."}` (shared_kernel httpapi), so the frontend never embeds the key.

### Fixes folded in
- `config/server.yaml` push notification `event_type: "execution_reminder"` → `"execution_ready_for_notification"` (the event the ExecutionWorker actually publishes; reminders never fire today).
- Deeplink templates `app://execution/%s` → `/maintenance/executions/%s` (the `app://` link is dead — the mobile app has no handler; web notification clicks will open the execution's page).

### DI / codegen
- Wire provider updates in `cmd/api/wire/` (`just wire`), mock regeneration (`just mock`).

## 2. Backend — completion field values

- `POST /v1/maintenance/executions/{id}/complete` body gains optional `field_values map[string]any` alongside `completed_by`.
- Values persist on the execution (`Execution.FieldValues` already exists in the domain and DTOs).
- Written test-first (repo TDD rule); OpenAPI spec (`docs/openapi.yaml`) updated.

## 3. Frontend — push enrollment

- **Service worker** `web/public/sw.js` (served from the app root by Vite's public dir):
  - `push` event → `self.registration.showNotification(title, {body, data: {deeplink}})`.
  - `notificationclick` → focus an existing client on the deeplink URL or `clients.openWindow` it.
- **Profile page** (`web/src/components/Profile.jsx`) gains a "Browser notifications" toggle:
  - On: fetch VAPID public key → register service worker → `Notification.requestPermission()` → `pushManager.subscribe({userVisibleOnly: true, applicationServerKey})` → `POST /v1/users/{userId}/push-tokens` with `{token: JSON.stringify(subscription), platform: "web"}`.
  - Off: `subscription.unsubscribe()` → `DELETE /v1/users/{userId}/push-tokens` with the same token string.
  - User ID comes from the auth context (`/v1/me`); permission-denied and unsupported-browser states render inline guidance.

## 4. Frontend — maintenance section (tenant portal)

New routes following existing component/CSS/api-helper patterns (`web/src/config/api.js`):

| Route | Page |
|---|---|
| `/portal/:tenantId/maintenance` | Activities list |
| `/portal/:tenantId/maintenance/new` | Create activity |
| `/portal/:tenantId/maintenance/activities/:activityId` | Activity detail (details/edit + history tabs) |
| `/portal/:tenantId/maintenance/up-next` | Up Next (pending executions) |
| `/maintenance/executions/:executionId` | Deeplink target — resolves the execution and redirects to its activity's history |

- **Activities list**: real pagination (`page`/`limit`), name, description, human-readable schedule, active/paused badge; actions: pause/resume (`POST .../activate|deactivate`), delete with confirmation (`DELETE`), create.
- **Create/Edit form**: type dropdown (`water_system`, `car`, `pets`, `custom` + required custom name), name, description, friendly cron builder (weekly/monthly/quarterly/yearly with day/interval dropdowns — JS port of the app's `schedule_utils.dart`, including reverse-parsing an existing cron with a fallback warning), reminder-days (`notification_days_before`) list editor, custom field definitions editor (name, display name, type `text`/`number`/`date`/`boolean`, required, default value). Edit sends only changed/allowed fields (type is read-only after creation, matching the server contract).
- **Up Next**: fetch activities (paged to completion), then executions per activity with parallel requests (`Promise.all`, not the app's sequential N+1); show earliest pending execution per activity sorted by date, relative-date labels, overdue badge with `overdue_days`, Complete button (hidden for future-dated executions, matching app behavior).
- **Activity detail — history tab**: executions sorted by `scheduled_date`, completed rows show `completed_at`, `completed_by`, and captured `field_values`.
- **Complete dialog**: `completed_by` auto-filled from the logged-in user's email; when the activity defines custom fields, inputs (typed per field definition, required flags enforced, defaults pre-filled) are captured into `field_values`.
- Cron format/parse helpers live in `web/src/utils/maintenanceSchedule.js`. The web project has no JS test runner (Vite build + ESLint only), so these helpers are validated through the functional test flows and manual verification; adding a JS test runner is out of scope.

## 5. Error handling

- Web push send failures: expired subscriptions self-clean (see §1); other failures log and count via the existing `zensor_server_push_notifications_total` metric.
- Frontend: API errors surface through the existing `NotificationSystem` toast pattern; permission-denied for notifications shows inline instructions rather than failing silently.

## 6. Testing

- **Unit (Ginkgo)**: `WebPushClient` against an `httptest` fake push service (success, 410 → `ErrSubscriptionExpired`), `PlatformRoutingPushClient` routing, worker platform propagation + expired-token cleanup, execution complete with `field_values`.
- **Functional (Godog, `@maintenance`)**: complete-with-field-values scenario; VAPID public key endpoint scenario.
- **Build gates**: `just unit`, `just functional maintenance`, `just build` (exercises the embedded web build), `just lint`, `just arch`.

## Out of scope

- Changes to the Flutter app (its FCM path keeps working unchanged).
- Web push for non-maintenance notifications beyond what already flows through the shared push pipeline (they benefit automatically via the routing client).
- Notification read/history UI in the web app.
