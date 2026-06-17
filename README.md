# Home Projects

A lightweight, self-hostable project manager for things happening around the
home — think "Jira, but simple". Users create **Projects**, fill them with
**Tasks** (and **Subtasks**), invite collaborators with roles, and optionally
share a project publicly via a read-only link.

The repo contains the **Go backend** (this directory) and a **Next.js +
shadcn/ui frontend** (`frontend/`).

## Concepts

- **Project** — a container of tasks. Private by default; can be made public
  (read-only to anyone with the link).
- **Task** — a unit of work with a status: `Waiting`, `In Progress`, `Done`,
  `Abandoned`. A **Subtask** is just a task with a `parentId`.
- **Project status** is *derived* from its tasks, never stored:
  - any task `In Progress` → project `In Progress`
  - all tasks `Abandoned` → project `Abandoned`
  - all tasks `Done`/`Abandoned` (≥1 Done) → project `Done`
  - otherwise → project `Waiting`
- **Members & roles** — `admin` (full control), `editor` (manage tasks),
  `viewer` (read-only).

## Architecture

| Concern        | Choice                                                            |
| -------------- | ---------------------------------------------------------------- |
| Language       | Go (stdlib + chi router)                                          |
| Database       | GORM with a **pluggable driver registry** — SQLite & PostgreSQL ship in-box; add more via `db.Register` |
| Auth (SSO)     | OIDC (e.g. **Pocket ID**); stateless JWT session cookie          |
| Authorization  | **OPA / Rego** policy evaluated **in-process** (`internal/authz/policy/authz.rego`) |
| Extensions     | Plugin registry — features register their own models & routes (see the 3D-printing extension) |
| Packaging      | Multi-stage, CGO-free build → distroless image                    |

```
cmd/server            entrypoint & graceful shutdown
internal/config       env-based configuration
internal/models       domain entities + status derivation
internal/db           pluggable GORM driver registry
internal/store        persistence (repository) layer
internal/auth         OIDC login, sessions, request principal
internal/authz        embedded OPA/Rego engine + policy
internal/api          HTTP server, routing, handlers
internal/extensions   plugin system
  └─ printing         3D-printing extension (external model links)
```

### Authorization model

Every API call is authenticated **except** login and reading public projects.
Each access decision is delegated to the Rego policy, which receives the action,
the caller (id + authenticated flag), and the project context (public flag,
owner, and the caller's role). The same `AuthorizeProject` check is exposed to
extensions so they reuse the core rules.

## Running locally

No external services required — defaults to on-disk SQLite and a dev-login that
stands in for SSO:

```bash
cp .env.example .env        # optional; defaults work out of the box
make run                    # listens on :8080
```

```bash
# Log in (dev mode), saving the session cookie
curl -c jar.txt -X POST localhost:8080/api/auth/dev-login \
  -H 'Content-Type: application/json' -d '{"email":"you@example.com","name":"You"}'

# Create a project
curl -b jar.txt -X POST localhost:8080/api/projects \
  -H 'Content-Type: application/json' -d '{"name":"Garage"}'
```

### With PostgreSQL

```bash
export HP_DB_DRIVER=postgres
export HP_DB_DSN="host=localhost user=hp password=hp dbname=homeprojects port=5432 sslmode=disable"
make run
```

### With Pocket ID (SSO)

Set `HP_OIDC_ISSUER`, `HP_OIDC_CLIENT_ID`, and `HP_OIDC_CLIENT_SECRET`. The
dev-login route is then disabled and the login flow lives at
`GET /api/auth/login` → `GET /api/auth/callback`.

## API overview

| Method & path                                   | Action            | Auth                |
| ----------------------------------------------- | ----------------- | ------------------- |
| `GET /healthz`                                  | health            | none                |
| `GET /api/auth/login` / `callback`              | SSO login         | none                |
| `POST /api/auth/dev-login`                      | dev login         | none (dev mode)     |
| `GET /api/auth/me`                              | current user      | required            |
| `POST /api/auth/logout`                         | logout            | —                   |
| `GET /api/projects`                             | list own projects | required            |
| `POST /api/projects`                            | create project    | required            |
| `GET /api/projects/{id}`                        | read project      | read (public ok)    |
| `PATCH /api/projects/{id}`                      | update project    | admin               |
| `DELETE /api/projects/{id}`                     | delete project    | admin               |
| `PATCH /api/projects/{id}/visibility`           | set public/private| admin               |
| `GET/POST /api/projects/{id}/members`           | members           | read / admin        |
| `DELETE /api/projects/{id}/members/{userID}`    | remove member     | admin               |
| `GET/POST /api/projects/{id}/tasks`             | tasks             | read / editor       |
| `PATCH/DELETE /api/projects/{id}/tasks/{taskID}`| task              | editor              |
| `GET/POST /api/ext/printing/projects/{id}/links`| print links       | read / editor       |

## Containers & Kubernetes

```bash
make docker                 # builds homeprojects:latest (distroless)
```

Sample manifests are in `deploy/k8s/`. The image is stateless aside from the
database; supply config via a ConfigMap and secrets via a Secret, and point
`HP_DB_DRIVER=postgres` at a managed PostgreSQL for production.

## Frontend

A Next.js (App Router) + TypeScript + Tailwind + shadcn/ui app lives in
`frontend/`. It talks to this API using cookie-based sessions
(`credentials: include`), so set `NEXT_PUBLIC_API_BASE_URL` to the backend URL
and make sure the backend's `HP_ALLOWED_ORIGINS` includes the frontend origin.

```bash
cd frontend
cp .env.local.example .env.local   # NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
npm install
npm run dev                        # http://localhost:3000
```

Pages: `/login` (SSO + dev-login), `/` (your projects), `/projects/[id]`
(tasks, subtasks, members, 3D-printing links), and `/p/[id]` (public read-only
view). Build the container with `docker build -t homeprojects-web ./frontend`.

## Tests

```bash
make test          # backend: status derivation + Rego policy
cd frontend && npm run lint && npm run build
```

CI (`.github/workflows/ci.yml`) runs both on every PR and publishes two images
to GHCR on push to `main` (datetime + `latest`) and on tag/release (the tag):
`homeprojects` (backend) and `homeprojects-web` (frontend).
