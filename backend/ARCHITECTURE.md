# Backend architecture

For the complete conventions and DDD / Clean / Hexagonal assessment, see
[`doc/architecture.md`](../doc/architecture.md),
[`doc/architecture-reference.md`](../doc/architecture-reference.md) and
[`doc/solid-ddd-hexagonal.md`](../doc/solid-ddd-hexagonal.md). Package dependency
rules are enforced by [`.go-arch-lint.yml`](.go-arch-lint.yml).

DDD / clean-architecture Go service. This document is the contract every module
follows so the codebase stays consistent as it grows.

## Layers and the dependency rule

```
infrastructure/http  ──►  application  ──►  domain
infrastructure/persistence ──►  application (ports) + domain
```

Dependencies point **inwards only**. Concretely:

| Layer | May import | Must NOT import |
|---|---|---|
| `domain/*` | `domain/shared`, `shared/apperrors` | `gorm`, `gin`, `net/http`, `infrastructure/*` |
| `application/*` | `domain/*`, `application/shared`, `shared/*` | `gin`, `net/http`, `infrastructure/*` |
| `infrastructure/http` | `application/*`, `domain/*` (types only), `httpx` | `infrastructure/persistence` |
| `infrastructure/persistence` | `application/*` (ports + read models), `domain/*` | `infrastructure/http` |

`shared/apperrors` is a transport-neutral error taxonomy (`Kind` + constructors).
It has **no** `net/http`. The `Kind → HTTP status` mapping lives only in
`infrastructure/http/httpx`.

## Write vs. read (CQRS-lite)

- **Write side** — `domain/<agg>/repository.go` defines a small `Repository`
  interface: `FindByID` (returns the aggregate's `ErrNotFound`, never `(nil,nil)`),
  `Save` (insert when `ID==0`, else update), `Delete`.
- **Read side** — `application/<agg>/read/queries.go` defines the read models
  (query projections, JSON-tagged — they are the HTTP response contract) and a
  `Queries` port (`List`, `Details`, …). `Details` returns `(nil, nil)` when
  absent; the use case turns that into `apperrors.NotFound`.
- Catalog / reference data (e.g. device brands & models) get their own port
  (`CatalogRepository`), not methods bolted onto the aggregate repository.

## Domain events + transactional outbox

1. Aggregates embed `shared.AggregateRoot` and call `Record(evt)` inside
   behaviour methods. Events embed `shared.BaseEvent` and implement
   `EventName() string`.
2. Write use cases run inside `TxManager.WithinTx`. Inside the closure:
   load → mutate via a domain method → `repo.Save` → `pub.Publish(ctx, agg.PullEvents()...)`.
3. `EventPublisher` is the outbox writer: it inserts the events into
   `outbox_messages` **in the same transaction**. Nothing is published unless the
   change commits.
4. `OutboxDispatcher` (started in the DI container) polls the table and
   re-publishes each row onto the in-process `events.Bus` as an `OutboxEnvelope`.
   Subscribe consumers (projections, notifications, …) in `di/container.go`.

### Delivery guarantees

- `events.Bus.Publish` is **synchronous** and returns the combined error of every
  handler; a panicking handler is recovered and counted as a failure.
- The dispatcher stamps `processed_at` **only when every handler succeeded**. On
  failure it records `last_error` and reschedules `next_attempt_at` with
  exponential backoff (2s → 5min). After `maxAttempts` (8) the row is
  **dead-lettered**: `failed_at` is set and it is no longer polled — inspect
  `SELECT * FROM outbox_messages WHERE failed_at IS NOT NULL`.
- At-least-once: handlers may see an event more than once (crash between publish
  and `processed_at`). Keep them idempotent.
- Single worker. If the process is scaled out, add `FOR UPDATE SKIP LOCKED` to
  the claim query before running multiple dispatchers.

### Consuming external events (inbox)

There is **no inbox component** — the service exposes a read-only external API
(`/api/ext/*`) but does not ingest events or webhooks from third parties. When
that need arises, mirror the outbox on the way in: a handler writes the raw
payload plus its provider message-id into an `inbox_messages` table with a
`UNIQUE` constraint on the id (dedupe), returns 2xx immediately, and a processor
drains unprocessed rows into use cases — the same poll/`processed_at`/backoff
shape as `OutboxDispatcher`.

## Rate limiting

`middleware.NewRateLimiter(redis, name, max, window)` is a fixed-window
per-client limiter. With Redis the window is shared across processes (`INCR` +
`EXPIRE NX` under key `ratelimit:<name>:<ip>`); without it — or if Redis errors
mid-request — it falls back to an in-process map so a cache outage never disables
limiting. Applied in `router.go` to `/auth/login`, `/auth/refresh` and
`/ext/auth/*`. Blocked requests get `429` + `Retry-After`.

## Idempotency keys

`middleware.IdempotencyKey(redis)` runs on the `/api` group and makes POST
requests safe to retry. It is **opt-in**: a client sends `Idempotency-Key: <uuid>`
and the first request executes while its response is cached (Redis, 24h; falls
back to an in-process store without Redis); any retry with the same key replays
the stored status/body and adds `Idempotency-Replayed: true`. A second request
that arrives while the first is still running gets `409`. 5xx responses are not
cached — the key is released so the client can retry cleanly. Only POST is
intercepted (PUT/PATCH/DELETE are idempotent by contract).

## Transactions

`application/shared.TxManager` is the only transaction abstraction. Its GORM
implementation stores the active `*gorm.DB` in the context; every repository
method resolves its connection through `connFromContext(ctx, r.db)`, so any repo
call inside `WithinTx` automatically enlists. Multi-step writes
(`CreateUser` + `AssignRoles`, token rotation, …) must be wrapped.

## RBAC

`middleware.Auth` validates the token, loads the caller's roles + effective
permissions (`appauth.ReadRepository.UserAuthz`) into the gin context, and flags
whether they hold a `FULL_ACCESS_ROLES` role. `middleware.RequirePermission("x")`
then gates each route; a full-access role bypasses the check.

Every `/api` route that isn't self-service (`/auth/*`, `/dashboard`) carries an
explicit permission in `router.go`: `<module>.{view,create,edit,delete}`. The
frontend guards the same permission names (`RequirePermission` component +
sidebar filtering), so the two layers stay in lock-step. Permissions are seeded
in `console/commands/seed.go` and assigned to the `admin` role.

## HTTP layer

- Request bodies bind into **application input structs** (`CreateInput`,
  `UpdateInput`, …). These carry `json` tags but **no** `binding` tags —
  required-field checks are explicit in the use case and return
  `apperrors.Invalid`.
- Responses are the `read` package types. Handlers never serialize a domain
  aggregate.
- Error handling: `httpx.Respond(c, err)` maps `apperrors.Kind` to a status code.

## How to add an aggregate `foo`

```
internal/domain/foo/
    entity.go        # aggregate + AggregateRoot + NewFoo(CreateData) + Update(UpdateData) + behaviour
    vo.go            # typed value objects (Estado, …) with Parse + predicates
    errors.go        # sentinel errors built from apperrors.* (carry a Kind)
    events.go        # FooCreated/Updated/... : BaseEvent + EventName + constructors
    repository.go    # write-side Repository interface (FindByID/Save/Delete)
internal/application/foo/
    foo.go           # UseCases struct + NewUseCases(...) + input DTOs + toDomain()
    write.go         # Create/Update/Delete/<behaviour> use cases (TxManager + EventPublisher)
    query.go         # List/Get use cases (delegate to read.Queries)
    read/queries.go  # read models (JSON contract) + Queries port
internal/infrastructure/persistence/
    foo_repository.go # FooRepository (write, tx-aware) + FooReadRepository (read) + mappers
internal/infrastructure/http/handlers/
    foo.go            # handler: bind input → use case → JSON(read model); httpx.Respond on error
```

Then wire one `foo.NewUseCases(...)` in `di/container.go` and register routes in
`infrastructure/http/router.go`.

## Verification

Run inside the container (`docker compose exec backend …`):

```
go build ./...
go vet ./...
go test ./...
golangci-lint run
make arch-lint
```

Integration tests (`-tags=integration`) need a real Postgres and **DELETE rows
on cleanup**, so they default to a dedicated `boilerplate_learner_test` database
(override with `TEST_DATABASE_DSN`). Create it once and run them:

```
docker compose -f compose.dev.yml exec postgres psql -U postgres -c 'CREATE DATABASE boilerplate_learner_test'
docker compose -f compose.dev.yml exec backend go test -tags=integration ./...
```
