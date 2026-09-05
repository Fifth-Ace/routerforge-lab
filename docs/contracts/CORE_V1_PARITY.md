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

---

## 14. Catalog package-management HTTP contract

Package lifecycle is exposed through two POST-only endpoints.

### POST /api/catalog/install

Request JSON:

```json
{
  "id": "<catalog-item-id>"
}
```

Behaviour:

- POST only
- same-origin required
- id must be non-empty
- operation maps to install
- small-JSON decoder limits size and rejects unknown fields

### POST /api/catalog/action

Request JSON:

```json
{
  "id": "<catalog-item-id>",
  "action": "install|update|remove",
  "confirm": "<optional typed confirmation>"
}
```

Behaviour:

- POST only
- same-origin required
- id must be non-empty
- action is trim+lowercase normalized
- only install/update/remove are executable
- HTTP execution context timeout is 180 seconds

## 15. Package-management execution gate

Package mutation is disabled unless at least one marker exists:

- /opt/etc/routerforge/package-management.enabled
- /opt/etc/dns-monitor/marketplace-test-install.enabled

Without a marker lifecycle execution fails closed with HTTP 403 semantics.

## 16. Catalog action authorization and serialization

All package lifecycle executions are serialized by a process-wide mutex.

Before execution Core verifies:

1. package management is enabled
2. catalog item exists
3. action is install/update/remove
4. catalog Actions explicitly allows that action
5. remove confirmation exactly equals item Name

Disallowed catalog action returns 403.
Remove confirmation mismatch returns 400.

## 17. Executable lifecycle methods

Only these methods execute:

- routerforge-release
- opkg
- structured

### routerforge-release

Requires release version, package, asset, SHA256 and URL.
Release package must match declared lifecycle packages when present.
Download is checksum verified before opkg execution.

Temporary directory:

/opt/tmp/routerforge-marketplace

Maximum downloaded asset size: 64 MiB.

Install:

opkg install <verified-local-ipk>

Update:

opkg --force-reinstall install <verified-local-ipk>

routerforge-release does not implement remove.

### Direct opkg

- install -> opkg install
- update -> opkg upgrade
- remove -> opkg remove

At least one package is required.

### Structured lifecycle

Supported steps:

- write-opkg-feed
- opkg-update
- opkg-install
- opkg-upgrade
- opkg-remove

Feed writes are restricted below /opt/etc/opkg/ and reject traversal-like paths.
Feed content must start with src/gz and contain an HTTPS URL.
Package names are restricted to the safe package-name character set.
The only explicit opkg argument currently allowed is --autoremove.
Ignored step failures remain visible in lifecycle output.

nfqws2 install refuses to proceed while legacy nfqws-keenetic is installed.

## 18. Lifecycle result and error contract

Successful result contains current concepts:

- id
- name
- action
- packages
- optional sources
- installed
- optional already_installed
- optional output
- completed_at

Returned command/output log is bounded to 16000 characters.
Lifecycle error detail is bounded to 4000 characters.

HTTP lifecycle errors contain:

- error
- detail
- result

A failed HTTP result does not prove that no package-manager side effect occurred.

## 19. Post-execution validation

After lifecycle execution Core rebuilds catalog state.

Install succeeds only if the item is subsequently detected as installed.
Remove succeeds only if the item is subsequently detected as not installed.
Update requires the item to remain installed and, when release version is known, installed version must match it.

opkg exit status zero alone is not sufficient proof of lifecycle success.

## 20. Current rollback semantics

The current Core lifecycle performs post-validation but DOES NOT implement a general automatic rollback transaction.

If opkg partially changes the system, a structured plan fails mid-way, or post-validation fails, Core reports failure but does not automatically restore the previous package/filesystem state.

Therefore:

- failure reporting is existing behaviour
- post-validation is mandatory existing behaviour
- general automatic rollback is currently absent

Adding rollback is a separate safety/product change, not a footprint-only optimization.

## 21. Admin proxy contract

Public namespace:

/api/admin/

Current Admin proxy is GET-only.
Non-GET methods return 405, Allow: GET and mutation_api=false.

/api/admin and /api/admin/ map to /v1/summary.
Other suffixes map to /v1/<suffix>.
Query parameters are preserved.

Unix sockets:

- /opt/var/run/routerforge-admin.sock
- /opt/var/run/dns-monitor-admin.sock

Upstream timeout is 6 seconds.
Upstream HTTP status is forwarded.
Responses use JSON content type and Cache-Control: no-store.

Unavailable Admin returns HTTP 503 with installed=false, running=false and mutation_api=false.

## 22. RouterForge release-index contract

Runtime release channel defaults to beta.
Recognized channels are beta and stable; unknown values normalize to beta.

Canonical repository:

Fifth-Ace/routerforge

Legacy fallback:

Fifth-Ace/dns-monitor

Cache path:

/opt/var/cache/routerforge/release-index-<channel>.json

Automatic refresh interval: 1 hour.
Remote HTTP timeout: 8 seconds.
Maximum release-index size: 512 KiB.

A failed remote refresh marks status offline/error but does not erase the currently held valid release document.

Validation includes:

- schema_version == 1
- channel matches selected channel
- at most 128 components
- valid package/version
- asset rejects traversal
- SHA256 exactly 64 hex characters
- release URL restricted to approved RouterForge GitHub release repositories
- duplicate package entries rejected

## 23. Phase 1 lifecycle safety conclusions

Frozen existing safeguards:

- explicit package-management gate
- same-origin mutation protection
- catalog-authorized actions
- typed remove confirmation
- serialized package execution
- allowlisted executable lifecycle methods
- checksum-verified RouterForge assets
- constrained structured opkg steps
- finite execution timeout
- post-execution catalog validation
- read-only Admin boundary
- stale valid release metadata survives remote refresh failure

NOT an existing guarantee:

- general transactional rollback

## 24. Phase 1 remaining work

Before Phase 1 is complete:

- freeze RouterForge registry/cache behaviour
- record package postinst/Core restart and inherited-FD behaviour
- enumerate parity test vectors
- create machine-runnable Core parity checks
- verify frozen contract against current Go implementation

No footprint optimization begins before those checks exist.
