# TinyCloud Milestones

TinyCloud validates one workflow:

> I built a small application. Give me a URL, let me inspect it, and let me share it.

The product starts as a TUI over one Docker host. Each milestone below is
intentionally sized to fit in one focused pull request. A milestone is complete
only when its acceptance criteria, tests, documentation, and manual demo are
included in the same PR.

## MVP boundary

The MVP is the first six milestones:

1. Local Docker deployment from the TUI
2. Reliable lifecycle and diagnostics
3. App configuration and reproducibility
4. Local sharing and access protection
5. HTTPS and public routing
6. A small control-plane boundary

Kubernetes, hosted multi-tenancy, billing, persistent databases, autoscaling,
and AI-assisted coding are deliberately outside the MVP. They should follow
evidence that the deploy-and-share workflow is useful.

## Delivery rules for every PR

- One user-visible capability or one enabling boundary per PR.
- Keep the diff reviewable: avoid unrelated refactors and speculative abstractions.
- Add or update unit tests for deterministic behavior and a manual demo checklist
  for runtime behavior.
- Update `README.md` or this file when the user workflow changes.
- Keep external effects behind a small interface so tests do not require Docker,
  a network, or a real browser.
- Make failure states explicit and recoverable; never report success before the
  runtime confirms it.
- Include the exact validation commands in the PR description.
- Prefer a vertical slice over completing an entire subsystem in isolation.

## Milestone 1 — Local Docker deployment (implemented)

**PR:** `feat/tui-docker-mvp`

**User outcome:** A developer can register a Dockerized project, deploy it to
the local Docker Engine, see its localhost URL, inspect logs, and stop it from
the TUI.

**Included:**

- Bubble Tea TUI with app list, filtering, and keyboard shortcuts.
- Create an app from a project directory and container port.
- Docker image build and container start through a runtime interface.
- Local JSON state under `.tinycloud/state.json`.
- App statuses: created, building, running, stopped, and failed.
- Log viewing, stop, URL display, and safe app-name normalization.
- Focused persistence, filtering, and slug tests.

**Acceptance criteria:**

- `go run ./cmd/tinycloud` starts the TUI.
- A Dockerfile project can be built and started with one deployment action.
- A failed build is shown as failed with its error retained in local state.
- Typing shortcut letters while filtering never triggers an app action.
- `go test ./...` passes without Docker.

**Out of scope:** HTTPS, authentication, remote hosts, databases, env vars,
health checks, and a web dashboard.

## Milestone 2 — Lifecycle safety and diagnostics

**PR:** `feat/runtime-health`

**User outcome:** A developer can tell whether an app is actually usable and
can recover from common failed or stale-container states.

**Included:**

- Docker inspect/status reconciliation on TUI startup and refresh.
- Explicit actions for deploy, restart, stop, and remove.
- Container exit code and last-known error display.
- Configurable command timeout and cancellation.
- Health-check polling with a small retry budget.
- Clear empty, unavailable, and Docker-not-installed states.

**Acceptance criteria:**

- Restarting TinyCloud does not falsely show an old app as running.
- A crashed container is distinguishable from a stopped container.
- A deployment timeout returns control to the TUI and preserves diagnostics.
- Runtime commands are unit-tested through a fake runtime.

**Out of scope:** application-defined health endpoints and resource limits.

## Milestone 3 — Reproducible app configuration

**PR:** `feat/app-manifest`

**User outcome:** The same app can be deployed again without re-entering its
configuration manually.

**Included:**

- Versioned `tinycloud.yaml` or `tinycloud.json` app manifest.
- Project directory, image name, container port, host port, and health path.
- `tinycloud deploy` as a non-interactive command using the manifest.
- TUI editing and validation for manifest values.
- Config validation errors before Docker is invoked.

**Acceptance criteria:**

- A checked-in manifest reproduces the same local deployment.
- Invalid ports, paths, names, and missing Dockerfiles fail early.
- TUI-created configuration can be inspected and edited by a user.
- Manifest parsing and validation are covered by table-driven tests.

**Out of scope:** secrets, environment values, automatic language detection,
and buildpacks.

## Milestone 4 — Local sharing with access protection

**PR:** `feat/local-share`

**User outcome:** A developer can give another person a temporary link to a
local app without exposing the Docker port directly.

**Included:**

- Local reverse proxy that routes `app.localhost` to the app port.
- Share action that creates a revocable, expiring access token.
- Lightweight access screen for protected apps.
- TUI actions to copy, revoke, and inspect a share link.
- Token hashing at rest and no token values in normal logs.

**Acceptance criteria:**

- A share URL reaches only the selected app.
- An expired or revoked token is rejected.
- Direct container ports can be disabled after proxy routing works.
- Proxy routing, expiry, and revocation are tested without a browser.

**Out of scope:** user accounts, email invitations, organizations, and roles.

## Milestone 5 — HTTPS and public routing

**PR:** `feat/https-routing`

**User outcome:** A developer can expose a selected app through a stable HTTPS
URL on a configured TinyCloud host.

**Included:**

- Configurable base domain and app subdomain generation.
- TLS termination through a standard reverse proxy.
- Automatic certificate provisioning in a development-safe mode.
- HTTP-to-HTTPS redirect and renewal/error status in the TUI.
- DNS and port prerequisites documented clearly.

**Acceptance criteria:**

- A configured domain routes to the intended container over HTTPS.
- Certificate failures are visible and do not claim a healthy deployment.
- App names cannot create invalid or conflicting hostnames.
- Routing logic has deterministic tests; live TLS is covered by a manual demo.

**Out of scope:** multi-region routing, wildcard hosting for untrusted tenants,
and production certificate operations across many hosts.

## Milestone 6 — Small control-plane boundary

**PR:** `feat/control-plane-api`

**User outcome:** The TUI can operate through a local daemon boundary, creating
the seam needed for a hosted TinyCloud later.

**Included:**

- Local HTTP API for apps, deployments, status, logs, and shares.
- TUI client implementation using the API instead of direct Docker calls.
- API request IDs, structured errors, and health endpoint.
- In-memory or file-backed repository behind the API.
- Contract tests using `httptest`.

**Acceptance criteria:**

- TUI behavior remains unchanged when switching from embedded to daemon mode.
- API errors are rendered as actionable TUI messages.
- No Docker implementation leaks into the TUI package.
- API tests run without Docker or network access.

**Out of scope:** PostgreSQL, OIDC, remote authentication, and deployment
queues.

## Post-MVP — Hosted platform increments

These milestones are intentionally deferred until the local workflow has users.
Each remains independently reviewable.

### Milestone 7 — Authentication and invitations

**PR:** `feat/oidc-invitations`

Add OIDC login, app ownership, email invitations, invitation expiry, and a
minimal owner/member permission model. Keep authorization decisions in the
control plane and cover them with table-driven tests.

### Milestone 8 — Persistent control-plane storage

**PR:** `feat/postgres-store`

Replace file state with PostgreSQL repositories, migrations, transactions, and
restart-safe deployment records. Keep the repository contract stable and add an
integration test profile separately from unit tests.

### Milestone 9 — Remote worker execution

**PR:** `feat/remote-worker`

Move Docker operations into a worker process with authenticated job submission,
deployment status events, bounded retries, and idempotent deployment IDs.

### Milestone 10 — Resource limits and observability

**PR:** `feat/resource-observability`

Apply CPU/memory/process limits, collect deployment duration and request/error
signals, expose structured logs, and add operator-facing failure diagnostics.

### Milestone 11 — Kubernetes runtime adapter

**PR:** `feat/kubernetes-runtime`

Implement the existing runtime boundary with Kubernetes Deployment, Service,
Ingress, ConfigMap, and Secret resources. The TUI and control-plane behavior
must remain unchanged; runtime-specific tests should use a fake client.

### Milestone 12 — Agent-native operations

**PR:** `feat/agent-api`

Expose safe, narrow operations for create app, deploy, status, logs, configure,
share, and destroy. Add idempotency, audit events, and confirmation for
destructive actions before considering MCP or other agent interfaces.

## Explicitly not planned for the first MVP

- General-purpose cloud hosting
- Arbitrary persistent databases
- Autoscaling and multi-region scheduling
- Billing and tenant-level quotas
- Built-in AI code generation
- A browser dashboard before the TUI workflow is validated
