# SOLID, DDD and Hexagonal Architecture

This guide applies to the Go backend (`backend/internal`). It records intended
conventions and distinguishes them from the current pragmatic implementation.

## DDD

### Bounded contexts and aggregates

The top-level domain packages—`auth`, `api_clients`, `organizations`, `curriculum`,
`cohorts`, `assessment`, `dispositivos`, `telefonos` and `trabajadores`—are business contexts with their own vocabulary. An aggregate should
protect invariants through behavior methods; other contexts should not mutate its
state directly. Coordinate cross-aggregate workflows in application code and pass
identifiers rather than reaching into another aggregate's internals.

Use value objects when they express a real constraint or domain concept. Avoid making
every primitive a wrapper without a domain reason.

### Repositories and domain events

Repository interfaces belong on the inside; GORM implementations belong under
`internal/infrastructure/persistence`. Domain events are raised by domain behavior,
stored transactionally with the write, and delivered at least once. Event handlers
with side effects must be idempotent.

## Hexagonal / Clean Architecture

```text
HTTP / console (inbound adapters) → application → domain
                                      ↑     ↓
                         persistence / cache / events (outbound adapters)
```

The application owns the contracts it needs. Infrastructure implements them. The
dependency-injection package selects concrete implementations at startup. Domain must
not know Gin, GORM, Redis, SQL, or HTTP status codes; application must not import
infrastructure packages.

The current learning use cases—including cohort management, enrollments, and student
lesson progress—are coordinated in `internal/application/learning`; cohort lifecycle
rules live in `internal/domain/cohorts`. Assessment rules live in `internal/domain/assessment`
and its use cases in `internal/application/assessment`. Persistence models and SQL
migrations remain adapters, not domain types.

The current auth use cases still call JWT and Argon2 libraries directly. This is a
known seam to improve, not a reason to add speculative abstractions elsewhere.

## SOLID in practice

- **Single responsibility:** handlers adapt HTTP; use cases coordinate; aggregates
  enforce invariants; repositories persist.
- **Open/closed:** add an adapter behind a useful port when another implementation is
  a real possibility, not for hypothetical extensibility.
- **Liskov substitution:** an adapter must honor the behavior and error semantics of
  its interface; test alternative/fake implementations against the same expectations.
- **Interface segregation:** keep ports small and use-case oriented; do not expose an
  entire ORM or framework through an interface.
- **Dependency inversion:** policies depend on inward contracts; adapters and the
  composition root depend on those policies.

## Adding a context

1. Define the business language, aggregate boundaries and invariants.
2. Add domain types, behavior and tests that do not need Gin or a database.
3. Define repository/read/transaction/event ports where the use case needs them.
4. Implement the use case in `internal/application/{context}`.
5. Add persistence/HTTP/other adapters under `internal/infrastructure` and wire them
   in `internal/di`.
6. Add integration tests for database mappings and HTTP contracts.
7. Update `backend/.go-arch-lint.yml` if a new package/layer is introduced; document
   exceptions in `doc/architecture-reference.md`.
