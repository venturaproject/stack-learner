<p align="center">
  <img src="doc/header.png" alt="Learner — multi-organization learning platform" width="100%" />
</p>

<h1 align="center">Learner</h1>

<p align="center">
  A learning platform for academies, training centers, and company teams.
</p>

<p align="center">
  <a href="https://venturaproject.github.io/stack-learner/">Project website</a> ·
  <a href="doc/README.md">Documentation</a> ·
  <a href="doc/product.md">Product</a> ·
  <a href="doc/architecture.md">Architecture</a> ·
  <a href="doc/architecture-reference.md">Technical guide</a> ·
  <a href="https://github.com/venturaproject/stack-learner">GitHub</a>
</p>

---

## What is Learner?

Learner is a **multi-organization LMS** for creating courses, organizing cohorts,
supporting student progress, and reviewing submissions. The backend and web application
share an API; authorization and organization boundaries are enforced by the server.

The project is under active development. See [doc/product.md](doc/product.md) for the
implemented scope and planned features.

## Features

- Organizations and memberships with administrator, instructor, and student roles.
- Courses with ordered modules and lessons, and draft/published states.
- Cohorts with assigned instructors, enrollments, and withdrawal history.
- Individual progress with sequential lesson unlocking.
- `task` activities: text and/or link submissions, attempts with history, a review queue,
  scores from 0 to 100, and feedback.
- Responsive web interface, available in Spanish and English.

Email invitations, account creation from the platform, and progress exports are not yet
available.

The project started from a technical template and still contains some inherited modules
outside the LMS scope; these will be replaced or removed over time.

## Stack

| Area | Technologies |
| --- | --- |
| API | Go 1.27 · Gin · GORM |
| Persistence | PostgreSQL 16 · SQL migrations |
| Cache and services | Redis 7 |
| Web | React · TypeScript · Vite · Tailwind CSS |
| Local environment | Docker Compose · Nginx |

## Get started

Requirements: Docker and Docker Compose.

Development containers do not restart automatically when Docker or your computer starts.
Start them explicitly with `make up`; stop them with `make down`.

```bash
make setup
make up
make migrate
make seed
```

Open [http://localhost:8092/login](http://localhost:8092/login). The local account
created by `make seed` is `admin@example.com` with password `admin1234` (**for development
only**). After signing in, open **Courses & Learning** and create an organization to get
started.

| Service | Local address |
| --- | --- |
| Application (Nginx proxy) | `http://localhost:8092` |
| Development frontend | `http://localhost:5174` |
| API | `http://localhost:8011/api` |
| API health check | `http://localhost:8011/api/health` |
| PostgreSQL | `localhost:5445` |
| Redis | `localhost:6380` |

`make up` applies pending migrations when the API starts; `make migrate` runs them
explicitly. Do not use local credentials in production, and change `JWT_SECRET` before
exposing the service.

## Checks

Run backend checks in the container:

```bash
make test
make test-integration
make vet
make staticcheck
make arch-lint
```

Frontend checks:

```bash
docker compose -f compose.dev.yml exec frontend pnpm build
docker compose -f compose.dev.yml exec frontend pnpm lint
docker compose -f compose.dev.yml exec frontend pnpm test
```

## Architecture and documentation

The backend is organized into modules with domain boundaries, use cases, and
HTTP/persistence adapters. Current details and decisions are documented here:

- [Product scope and permissions](doc/product.md)
- [Backend architecture](doc/architecture.md)
- [Architecture reference and current trade-offs](doc/architecture-reference.md)
- [Frontend architecture and Clean Architecture conventions](doc/frontend-architecture.md)
- [DDD, hexagonal architecture, and SOLID](doc/solid-ddd-hexagonal.md)

## License

A license has not yet been selected for this repository.
