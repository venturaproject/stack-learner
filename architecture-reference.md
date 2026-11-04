# Architecture: technical reference

This is the implementation-specific convention guide for this Gin + Go boilerplate.
The Spring boilerplate's `doc/` is a format/reference only; its Java, Spring, JPA and
ArchUnit rules do not apply here.

Current backend baseline (checked 2026-10-27): Go 1.27.3 toolchain (`go 1.27.0` in
`go.mod`) and Gin v1.12.0. `go list -m -u github.com/gin-gonic/gin` reports no newer
module version; keep Gin pinned at the current version until an update is available
and tested.

## Package responsibilities

| Package | Responsibility | Dependency policy |
|---|---|---|
| `internal/domain/{context}` | Entities/aggregates, value objects, invariants, repository contracts and domain events | May use `internal/domain` and neutral `internal/shared`; never application, HTTP, GORM or infrastructure |
| `internal/application/{context}` | Use cases, input/output models, read ports, transaction/event ports | May use domain and neutral shared code; never infrastructure, Gin, `net/http` or config loaders |
| `internal/infrastructure/http` | Gin router, middleware, handlers, HTTP response/error mapping and request pagination | Inbound adapter; may call application and use domain types at the boundary |
| `internal/infrastructure/persistence` | GORM repositories, persistence models, transaction manager and outbox writer | Outbound adapter; implements inward-defined ports |
| `internal/infrastructure/{cache,events,console}` | Redis, event bus and CLI adapters | May use inward contracts and config; no dependencies on HTTP |
| `internal/di` | Concrete dependency wiring | Explicit exception: may import all application and infrastructure components |
| `cmd/*` | Process entry points | Start composition/wiring; avoid application business logic |

`internal/infrastructure/http/pagination` deliberately lives at the HTTP edge because
it parses `gin.Context`; it is not a general-purpose shared/domain package.

## DDD conventions

- Model one business context per `internal/domain/{context}` package.
- Keep business invariants in domain methods/constructors, not handlers or GORM hooks.
- Use intention-revealing methods for state changes. Avoid exporting mutable state
  solely to make persistence convenient.
- Define repository and other required capabilities as inward-facing interfaces;
  implement them in infrastructure.
- Reference other aggregates by identity and coordinate multi-aggregate work in an
  application use case.
- Keep read projections separate from aggregate write contracts where practical.
- Domain events express business facts. Persist event records in the same transaction
  as the state change; consumers must tolerate at-least-once delivery.
- Add abstractions when they protect an invariant, a boundary or testability—not just
  to satisfy a pattern checklist.

## Ports, adapters and transactions

Application ports currently include repository/read interfaces, `TxManager` and event
publishing. GORM, Gin, Redis and the SQL outbox are adapters. `internal/di/container.go`
is the only intentional layer-crossing composition point.

Write use cases that mutate more than one record should use the transaction port.
Adapters must keep the port's semantics; tests should exercise use cases without a
live database where possible and persistence mappings against PostgreSQL for
integration coverage.

Organization-scoped learning use cases enforce membership roles and student course
visibility in the application service. Role checks and the associated reads/writes run
through the transaction port; persistence scopes queries to the organization. Some
data-backed prerequisites (such as active enrollment and sequential lesson unlocking)
are currently checked by repository methods because they are evaluated alongside
course/progress records; this is a placement compromise, not a policy for new use cases.

Assessment use cases follow the same boundary: the application enforces membership roles
and reviewer assignment, while a repository check verifies the task kind, published course,
active enrollment, and sequential unlock before student submission/history access.
Organization admins and course-assigned trainers may list and review submissions.
Submission creation and review run inside the transaction port; persistence scopes every
read and write by organization and course.

Organization creation also uses the application transaction port because it writes both
the organization and its creator's `org_admin` membership. Course edits/deletes are
restricted to drafts; published courses cannot be changed or deleted through the course
endpoints.

## Explicit compromises / items to improve

This is pragmatic DDD and hexagonal architecture, not a framework-free Clean
Architecture implementation:

1. The application auth package currently performs Argon2 password hashing and JWT
   creation/validation. These are technology-specific concerns that could move behind
   application ports with infrastructure adapters.
2. `internal/shared/apperrors` is intentionally imported by domain and adapters so
   transport-neutral error kinds can cross boundaries. It must remain free of Gin and
   `net/http` status codes.
3. The console's seed/user commands can use persistence models directly for bootstrap
   operations; new business workflows should call application use cases instead.
4. The architecture linter checks project-package dependencies. It currently allows
   third-party imports broadly; layer-specific third-party restrictions are not yet
   enforced. Keep framework dependencies out of domain/application unless a documented
   exception is added.

These are documented exceptions, not evidence that the service is fully framework
independent. Move them incrementally when there is a concrete benefit and update the
architecture config/tests in the same change.

## Executable conventions

Database schema changes are tracked in `backend/internal/infrastructure/persistence/migrations/`
as monotonically numbered SQL files. `console migrate` applies pending migrations under a
PostgreSQL advisory transaction lock and records them in `schema_migrations`; application
startup invokes the same idempotent runner as a safety net. Databases predating this
history receive a one-time GORM baseline, followed by the numbered SQL migrations. Do not
add new `AutoMigrate` calls outside that baseline. Reference data belongs in migrations;
development/demo data belongs in idempotent `console seed` seeders. `seed --fresh` is
destructive and must only target a disposable development database.

`backend/.go-arch-lint.yml` classifies packages and blocks outward layer imports.
Run it with:

```sh
make arch-lint
```

The same check can be run directly with the pinned analyzer image:

```sh
docker run --rm -v "$PWD/backend:/app" \
  fe3dback/go-arch-lint:release-v1.19.0 check --project-path /app
```

When introducing a new package, assign it to a component and state its permitted
dependencies. A rule exception should be narrow, explained here, and accompanied by a
test/architecture-config change.
