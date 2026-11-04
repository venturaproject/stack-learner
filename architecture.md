# Backend architecture

This describes the Go backend as implemented. The executable import rules live in
[`backend/.go-arch-lint.yml`](../backend/.go-arch-lint.yml); this document explains
their intent and records current trade-offs.

## Layers

```text
cmd / infrastructure (HTTP, console, persistence, cache, events)
                         │
                         ▼
                    application
                         │
                         ▼
                      domain
```

- `internal/domain/{context}` contains aggregates/entities, value objects, domain
  errors/events and repository contracts.
- `internal/application/{context}` coordinates use cases and declares the ports it
  consumes. Read projections live under each context's `read` package.
- `internal/infrastructure` implements persistence, HTTP, console, cache and event
  adapters. `internal/di` is the composition root and is intentionally allowed to
  know all layers.
- `internal/config`, `internal/shared` and `pkg/logger` contain cross-cutting
  configuration, transport-neutral errors and logging.
- `cmd/server`, `cmd/console` and `cmd/seed` are entry points.

The allowed project-package dependencies point inward: domain does not import
application or infrastructure; application does not import infrastructure; HTTP and
persistence adapters implement application/domain contracts. The composition root
assembles concrete adapters.

## Request and persistence flow

```text
HTTP request → infrastructure/http/handlers → application use case
             → domain rules + ports ← infrastructure/persistence
```

Handlers translate the HTTP request and response. Application code coordinates a use
case; aggregates own their business invariants. Persistence maps between domain types
and GORM models. Transaction boundaries are expressed through the application
`TxManager` port and implemented by persistence.

## Related references

- [Technical architecture and conventions](architecture-reference.md)
- [DDD, Clean/Hexagonal architecture and SOLID](solid-ddd-hexagonal.md)
- [Existing backend architecture notes](../backend/ARCHITECTURE.md)
