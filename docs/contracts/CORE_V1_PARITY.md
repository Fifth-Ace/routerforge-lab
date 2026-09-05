# RouterForge Core v1 contract freeze

Status: Phase 1 baseline
Baseline commit: f394ecae7621ca8cd14bde6c4832a7dece6817ff
Scope: externally observable Core behaviour that must survive optimization or reimplementation.

This document is a compatibility contract, not a design proposal.

## 1. Core health

### GET /api/health

Public even when authentication is enabled.

Response is JSON and includes:

- `ok: true`
- `version`
- `module_abi: "v1"`

Headers:

- `Content-Type: application/json`
- `Cache-Control: no-store`

This endpoint is used by package/runtime health checks and must remain cheap and dependency-independent.

---

## 2. Authentication

Public auth endpoints:

- `GET /api/auth/status`
- `POST /api/auth/login`
- `POST /api/auth/logout`
- `POST /api/auth/config`
- `GET /api/health`

Additionally, loopback-only `GET`/`HEAD` module health requests may bypass normal API authentication.

### Security invariants

- Security config path:
  `/opt/etc/routerforge/security.json`
- Missing config means authentication is not required.
- Existing unreadable or malformed config fails closed:
  authentication becomes required.
- Session cookie name:
  `routerforge_session`
- Session TTL:
  12 hours.
- Cookie:
  - path `/`
  - HttpOnly
  - SameSite=Strict
- Login and auth configuration mutation reject cross-origin requests.
- Enabling authentication requires valid Entware root credentials.
- Disabling authentication while enabled requires an authenticated session.
- Authentication backend identity remains `entware-root`.
- Only username `root` is accepted by the current backend.
- Request JSON is size-limited and rejects unknown fields.

### Login throttling

Per client:

- failure window: 5 minutes
- 5 failed attempts triggers blocking
- block duration: 30 seconds
- blocked response uses HTTP 429 and `Retry-After`

A replacement Core must not silently weaken any of these rules.

---

## 3. Snapshot API

### GET /api/snapshot

Read-only.

Headers:

- JSON content type
- `Cache-Control: no-store`

Core returns its own version/server time and merges the DNS module `/v1/snapshot` payload when DNS is available.

Guaranteed Core-level fields include:

- `version`
- `server_time`
- `dns_module_online`

When DNS is unavailable:

- request still succeeds
- `dns_module_online` is false

When DNS is available:

- DNS snapshot fields are merged
- Core's own version and server time win
- `dns_module_online` is true

DNS failure must not make Core snapshot unavailable.

---

## 4. Server-Sent Events

### GET /api/events

Content type:

`text/event-stream`

Headers include:

- `Cache-Control: no-cache, no-transform`
- `Connection: keep-alive`
- `X-Accel-Buffering: no`

Initial retry directive:

`retry: 1500`

Event name:

`snapshot`

Default interval:

2 seconds.

Optional query:

`interval_ms`

Clamp:

- minimum 1000 ms
- maximum 30000 ms

Invalid values fall back to default behaviour.

Each event payload uses the same Core snapshot semantics as `/api/snapshot`.

### Authentication invariant

If authentication becomes required, an anonymous SSE connection must not continue receiving snapshots indefinitely.

The current implementation checks authentication before each send and closes an unauthenticated stream.

---

## 5. Legacy DNS compatibility routes

The following Core routes are compatibility aliases into the DNS module and must remain functional until explicitly versioned/deprecated:

- `/api/history`
- `/api/quality`
- `/api/fallbacks`
- `/api/error-bursts`
- `/api/clients`
- `/api/client`
- `/api/interfaces`
- `/api/system`
- `/api/plain-dns`
- `/api/dns/info`

They proxy into:

`/api/modules/dns/<target>`

A Core optimization must not remove these merely because the frontend no longer uses one of them.

---

## 6. Module proxy ABI

Public route namespace:

`/api/modules/<module>/...`

Known socket mappings:

- dns:
  `/opt/var/run/routerforge-dns.sock`
- system:
  `/opt/var/run/routerforge-system.sock`
  fallback `/opt/var/run/dns-monitor-system.sock`
- thermal:
  `/opt/var/run/routerforge-thermal.sock`
  fallback `/opt/var/run/dns-monitor-thermal.sock`
- storage:
  `/opt/var/run/routerforge-storage.sock`
  fallback `/opt/var/run/dns-monitor-storage.sock`
- network:
  `/opt/var/run/routerforge-network.sock`
  fallback `/opt/var/run/dns-monitor-network.sock`

### Method policy

DNS:

- GET
- HEAD
- POST
- PATCH
- DELETE

System/Thermal/Storage/Network:

- GET
- HEAD only

Profiling:

- GET
- HEAD only
- status-only pseudo-module behaviour

A replacement implementation must preserve the distinction between mutable DNS and read-only monitoring modules.

### Path translation

External:

`/api/modules/<module>/<rest>`

Internal module API:

`/v1/<rest>`

Empty rest maps to:

`/v1/health`

Traversal-like paths containing `..` are rejected.

Trailing slash behaviour for module UI paths must be preserved so internal `/v1` redirects are not leaked to the browser.

---

## 7. Module failure behaviour

Unknown module:

HTTP 404.

Known but unavailable API module:

HTTP 503 JSON.

Failure response includes current concepts:

- module
- installed
- running=false
- mutation_api
- error
- detail

Module UI failure is intentionally different.

For `/v1/ui` paths Core returns a reconnect HTML page with HTTP 503, including:

- `Retry-After: 1`
- `Cache-Control: no-store`
- `X-Content-Type-Options: nosniff`

The reconnect page polls module health and reloads when the module becomes available.

This behaviour is part of the user-visible module restart experience.

---

## 8. Catalog API

### GET /api/catalog

Read-only catalog snapshot.

Headers:

- JSON
- `Cache-Control: no-store`

### POST /api/catalog/refresh

Forces RouterForge release index and registry refresh concurrently, then returns:

- `ok`
- release status
- registry status
- resulting catalog

Mutation method remains POST.

### Catalog lifecycle endpoints

Current namespaces that must retain behaviour:

- `/api/catalog/action`
- `/api/catalog/install`

Their detailed request/rollback/post-validation contract is frozen separately in the package-lifecycle section of Phase 1 before implementation work begins.

---

## 9. Admin proxy

Namespace:

`/api/admin/`

Core remains the HTTP-facing gateway for Admin.

Detailed Admin route/method/error parity will be enumerated separately before Core implementation changes.

No optimization may silently bypass Core authentication or expose Admin directly.

---

## 10. SPA/frontend serving

Root namespace serves the RouterForge SPA.

Allowed methods for frontend content:

- GET
- HEAD

Other methods do not fall through to SPA rendering.

### Index

`index.html`:

- `Content-Type: text/html; charset=utf-8`
- `Cache-Control: no-cache`

Unknown frontend routes fall back to index for SPA routing.

### Immutable assets

Files below:

`_app/immutable/`

receive:

`Cache-Control: public, max-age=31536000, immutable`

Any frontend packaging optimization must preserve SPA fallback and cache semantics.

---

## 11. Global HTTP middleware

Authentication middleware wraps API access.

Profiling middleware wraps the HTTP handler when enabled.

Core HTTP server currently has a 5-second `ReadHeaderTimeout`.

Externally visible status codes, auth boundaries, caching semantics and streaming behaviour are compatibility-sensitive even if implementation language changes.

---

## 12. Non-negotiable migration rule

Optimization/native work is not considered parity-complete merely because the UI opens.

The candidate Core must be checked against:

1. endpoint existence
2. HTTP methods
3. status codes
4. response JSON shape
5. headers/cache policy
6. authentication boundary
7. module method policy
8. Unix-socket routing
9. unavailable-module behaviour
10. SSE streaming behaviour
11. SPA fallback/cache behaviour
12. package lifecycle safety

Any intentional contract change must be isolated, documented and reviewed separately from footprint optimization.

---

## 13. Phase 1 remaining contract work

Still to freeze before Phase 1 is complete:

- Admin proxy route/error details
- catalog action/install request schema
- opkg execution lifecycle
- install/update/remove validation
- rollback and post-validation semantics
- release/registry cache and stale-data behaviour
- package postinst/Core restart lifecycle
- automated parity test vectors

No Core optimization begins until these sections are frozen.
