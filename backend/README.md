# Backend — Go

Go 1.26 / **Clean Architecture** / Gin for routing and middleware.

## Running

```bash
# Docker (from the repository root)
make up
make exec-backend

# Directly on the host
make start              # :4000
make port=5555 start
make dev                # hot reload (requires air)

# Smoke check
curl localhost:4000     # {"message":"pong"}
```

Installing air:

```bash
go install github.com/air-verse/air@latest
```

## API documentation (Swagger / OpenAPI)

The HTTP surface is documented with swag annotations in `apispec/*.go` and served
through a Swagger UI. This covers the session-cookie endpoints as well as the
**implemented Public API** (`/v1/*`, Bearer API keys) and key management (`/apikeys`).

```bash
make swagger          # regenerate the spec from apispec and cmd/swaggo/main.go
make swagger-serve    # regenerate and start the Swagger UI
open http://localhost:4000/swagger/index.html
```

| Thing                         | Location                                           |
| ----------------------------- | -------------------------------------------------- |
| Hand-written spec source      | `apispec/*.go`                                     |
| General info + UI server      | `cmd/swaggo/main.go`                               |
| Generated output (don't edit) | `docs/swagger/{docs.go,swagger.json,swagger.yaml}` |
| Spec (OpenAPI 2.0)            | `http://localhost:4000/swagger/doc.json`           |

Conventions:

- Request/response DTOs never expose `infrastructure` models directly, so
  `password_hash`, `token_hash` and `storage_key` can never leak.
- `UserPublic` (for others) and `UserPrivate` (for yourself) are separate types;
  email is only returned to its owner.
- List responses share `limit` / `offset` / `total` (`Pagination`).
- Failures carry a machine-readable `code` where the client needs to translate
  them (`writeJSONErrorCode`); the client renders localized text keyed by code.
- After changing annotations, run `make swagger` and commit the regenerated files.
- Some `apispec` entries (e.g. `/auth/refresh`, `/blocks`) are specification-first
  and not implemented yet; the implemented routes are registered in
  `handler/handler.go`.

## Architecture

### The one rule: dependencies point inward

```
outer   handler                    infrastructure
        HTTP in/out (Gin routes)    DB, external services
           │                          │
           │ calls                    │ implements interfaces
           ▼                          ▼
inner   usecase   orchestration + repository interface declarations
           │
           ▼
inner   domain    entities and rules; depends on nothing
```

The outer layers (technical detail: DB, HTTP) may know the inner ones, but the
**inner layers must never know the outer ones**. `domain` depends on nothing.

`handler` and `infrastructure` are **siblings that never import each other**; the
composition root (`cmd/serv/main.go`) wires them together.

Why:

- Swapping PostgreSQL for something else leaves `domain` and `usecase` untouched.
- Swapping HTTP for WebSocket/gRPC keeps the business logic as-is (our game does
  exactly this: the same usecase serves both).
- `usecase` tests inject in-memory fakes instead of a real database.

> The original book calls the handler layer "interface adapters"; `interface` is a
> Go keyword and makes import paths noisy, so we use `handler/`.

Note on Gin: routes and middleware live at the edge only. Handlers keep the
`net/http` signature and are mounted through a small `wrapF` adapter
(`handler/handler.go`) that copies `c.Param` into `r.PathValue`, so the inner
layers never see Gin.

### Layer responsibilities

| Layer          | Directory         | Responsibility                                   | May depend on    |
| -------------- | ----------------- | ------------------------------------------------ | ---------------- |
| domain         | `domain/`         | Entities and business rules                      | nothing          |
| usecase        | `usecase/`        | Flow of operations; declares needed interfaces   | domain           |
| handler        | `handler/`        | HTTP ⇔ domain translation, routing               | usecase, domain  |
| infrastructure | `infrastructure/` | Technical detail; implements usecase interfaces  | usecase, domain  |

### The hinge: dependency inversion

The arrows point inward, yet `usecase` needs the database. The resolution:
**usecase declares what it needs as an interface, on its own side.**

```
      usecase package
      ┌──────────────────────────────┐
      │ type GameRepository          │  ← "this is what I need" (abstract)
      │     interface { ... }        │
      │                              │
      │ GameUsecase works against    │
      │ that interface only          │
      └──────────────────────────────┘
                    ▲
                    │ implements (dependency still points outer → inner)
      ┌──────────────────────────────┐
      │ infrastructure.GameRepo      │  ← actually talks to PostgreSQL (concrete)
      └──────────────────────────────┘
```

`usecase` only knows "something that can persist matches". Whether that is
PostgreSQL or an in-memory fake is decided in the composition root.

### Adding a feature

Build **from the inside out**:

1. `domain/xxx.go` — entities and rules (no HTTP, no DB)
2. `usecase/xxx.go` — the flow; declare dependencies as `XxxRepository` interfaces
3. `infrastructure/xxx.go` — implement the interfaces (an in-memory fake first is fine)
4. `handler/xxx.go` — decode JSON, call the usecase, encode the result
5. `handler/handler.go` + `cmd/serv/main.go` — register the route, inject concretes

Review checklist: [pull_request_template.md](../.github/pull_request_template.md).

## Database / migrations

The authoritative, always-current schema documentation is generated with tbls into
[docs/schema/](../docs/schema/README.md) (per-table Markdown + ER diagrams). The
pipeline from code to database:

```
infrastructure/model.go          GORM structs (annotated with invariants)
        │  go run ./cmd/migrate
        ▼
DDL (CREATE TABLE ...)           the desired schema
        │  atlas migrate diff
        ▼
migrations/*.sql                 versioned migration steps
        │  atlas migrate apply
        ▼
PostgreSQL
```

`atlas migrate apply` **runs automatically when the container starts**
([docker-entrypoint.sh](docker-entrypoint.sh)); the Atlas CLI ships in the image,
so `make up` applies any pending migration. The commands below are for **creating**
migrations or running against the host directly:

```bash
make schema                        # print the DDL (no DB needed)
make migrate-diff name=add_users   # generate migrations/*.sql from the diff
make migrate-apply                 # apply manually

curl -sSf https://atlasgo.sh | sh  # install Atlas on the host
```

Things to know:

- **GORM `AutoMigrate` is never used.** It silently rewrites tables at boot with
  no record and no rollback. Atlas keeps every change as a reviewed SQL file.
- **`make schema` output is not the final schema.** Partial unique indexes,
  `CREATE EXTENSION citext`, multi-column CHECKs and circular FKs can't be
  expressed in GORM tags — they are added as hand-written SQL in `migrations/`.
  Each struct's doc comment in `infrastructure/model.go` lists what to add
  (`@migration` lines). When `atlas migrate diff` proposes DROPs for those
  hand-written constraints, delete the DROPs from the generated file and rehash.
- **GORM models live in `infrastructure`, not `domain`.** GORM tags are database
  detail; putting them in `domain` would break "depends on nothing". Conversion
  between the two is the repository's job.
- `cmd/migrate` only prints DDL to stdout; it migrates nothing by itself. See the
  comment atop [atlas.hcl](atlas.hcl) for the configuration details.
- If migrations were skipped you get `relation "..." does not exist (SQLSTATE 42P01)`
  — the entrypoint normally prevents this, even after `make fclean`.

## Dependency management (go.mod / go.sum)

| File     | Role                                                                    | Analogy                    |
| -------- | ----------------------------------------------------------------------- | -------------------------- |
| `go.mod` | Module name, Go version, **declared** dependencies and versions         | Shopping list              |
| `go.sum` | **Hash ledger** of every downloaded module; detects tampering           | Receipt + tamper seal      |

**Both are committed.** Anyone building at any time gets identical dependencies.

`// indirect` marks modules our code never imports directly. Most of the ~1700
`go.sum` lines come from `atlas-provider-gorm` dragging in every SQL driver — used
only by `cmd/migrate`, never by the app itself.

| Command         | What it does                                                        |
| --------------- | ------------------------------------------------------------------- |
| `go get X`      | Add/update X                                                        |
| `go mod tidy`   | Reconcile with actual imports. **Run after every dependency change** |
| `go mod verify` | Check the local cache against `go.sum`                              |

## Lint / Format

Formatting and static analysis are unified under **golangci-lint v2**
(enabled linters: [.golangci.yml](.golangci.yml)).

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
```

| Command          | What it does                                |
| ---------------- | ------------------------------------------- |
| `make lint`      | Static analysis (same as CI)                |
| `make fmt`       | Format (gofmt + goimports)                  |
| `make fmt-check` | Detect unformatted files (diff only)        |
| `make ci`        | `build` + `lint` + `fmt-check`              |
| `make build`     | Build to `./tmp/main`                       |

For a justified one-off exception, annotate the line:

```go
//nolint:errcheck // best-effort; keep going on failure
```

CI: [.github/workflows/backend-lint-format.yml](../.github/workflows/backend-lint-format.yml)
runs on PRs/pushes touching `backend/**`. Run `make ci` before pushing.

## Environment variables

| Variable                | Description                                       | Default                                        |
| ----------------------- | ------------------------------------------------- | ---------------------------------------------- |
| `PORT`                  | Listen port                                       | `4000`                                         |
| `DATABASE_URL`          | PostgreSQL connection string                      | –                                              |
| `GOOGLE_CLIENT_ID`      | Google OAuth client ID                            | – (required)                                   |
| `GOOGLE_CLIENT_SECRET`  | Google OAuth client secret                        | – (required)                                   |
| `GOOGLE_REDIRECT_URL`   | OAuth callback (must match the Console entry)     | `https://localhost:8443/auth/google/callback`  |
| `FRONTEND_URL`          | Where to send the browser after login             | `https://localhost`                            |
| `MEDIA_DIR`             | Directory for uploaded avatars                    | `./uploads`                                    |
| `PUBLIC_API_RATE_LIMIT` | Public API requests per minute per key            | `60`                                           |

When running via Docker these come from the root `.env` / compose defaults.
