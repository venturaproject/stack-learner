# Frontend architecture and design conventions

This document sets the frontend's current conventions and the direction for applying
Clean Architecture in Learner. It is grounded in the existing React application; it is
not a claim that the frontend has already been fully refactored into these boundaries.
Adopt the rules incrementally when adding or changing a feature. Avoid broad moves or
abstractions without a concrete use case.

## Goals

- Keep product rules independent of React, routing, HTTP, and visual components.
- Make organization scope, roles, loading, empty, and error states explicit.
- Let UI code express user intent without owning transport details or duplicating
  server-side rules.
- Keep features testable without rendering the whole application or starting the API.
- Prefer small, cohesive feature modules over a large generic framework.

## Current application map

The current `frontend/src` structure includes:

| Location | Responsibility |
| --- | --- |
| `app.tsx`, `app-routes.tsx`, `main.tsx` | Application startup, route registry, and lazy route loading. |
| `routes/` | Route-level decisions and route parameters, such as resolving a course before editing. |
| `pages/` | Screen composition and page-specific interaction. Learner screens currently live under `pages/learning`, `pages/cohorts`, and related folders. |
| `components/`, `layouts/` | Shared UI, navigation, and page shells. |
| `services/` | HTTP-facing API functions. For example, `services/learning-api.ts` uses the shared Axios instance. |
| `lib/` | Cross-cutting browser/application utilities and pure helpers. Keep product policies here only while they are small and clearly named; move feature rules into the feature when they grow. |
| `hooks/`, `context/` | Reusable React behavior and genuinely shared client state. |
| `schemas/`, `types/` | Validation schemas and shared TypeScript types. Prefer feature-local definitions when they are not shared. |
| `config/`, `i18n/` | Environment/configuration and localization infrastructure. |

This is a transitional structure inherited in part from the starter template. Some
unrelated legacy modules remain. Do not use those modules as the design reference for
new Learner features.

## Dependency rule

Dependencies point inward: presentation depends on application behavior; application
behavior depends on domain rules and declared capabilities; infrastructure implements
those capabilities. Domain rules must not depend on React, Axios, TanStack Query, or
browser APIs.

```text
React routes / pages / components
              │ user intent and view state
              ▼
Application hooks / use cases ───────► Domain policies and models
              │                             (pure TypeScript)
              ▼
Infrastructure adapters ─────────────► HTTP API, storage, browser
```

The arrow describes allowed knowledge, not a requirement to create one class or folder
per box. In a small feature, a query hook may call a focused API adapter directly. Add
an application service or port when it protects a meaningful rule, coordinates multiple
operations, or makes the feature easier to test—not just to wrap one function.

### Suggested feature shape for new or substantially changed modules

Keep files close to the feature and introduce only the boundaries it needs:

```text
src/features/learning/
  domain/          # Course policies and transport-independent types
  application/     # Queries, commands, orchestration, query keys
  infrastructure/  # HTTP adapter and DTO mapping
  presentation/    # Routes, pages, feature components, and forms
```

The current `pages/learning` and `services/learning-api.ts` layout is valid legacy code.
Do not make a second competing implementation. When extending it, keep each new file
cohesive and migrate a boundary only when there is a clear benefit. Shared UI stays in
`components/`; feature-only UI stays with the feature.

## Responsibility conventions

### Domain rules

- Put deterministic product rules in plain TypeScript functions, types, or small value
  objects. Keep them free of React and I/O.
- Keep authorization and workflow decisions named and testable. A UI permission helper
  can hide or disable an action for usability; the API remains the source of truth and
  must enforce every permission.
- Keep organization isolation server-enforced. Treat the selected `organization_id` as
  untrusted input; never infer authorization from a selector, route guard, or hidden
  button.
- Model only distinctions the product actually supports. Do not invent front-end states
  that the API cannot represent.

### Application behavior and server state

- Use TanStack Query for remote/server state. Do not mirror query data into `useState`
  unless the user is editing a separate draft.
- Centralize query keys and feature query/mutation behavior as it becomes shared or
  non-trivial. Include all scope and filter inputs in query keys (organization, entity,
  page, search, and similar values).
- Gate queries until required IDs and permissions are available. Avoid non-null
  assertions as a substitute for a valid query precondition.
- Mutations should express a user command, handle success/error consistently, and
  invalidate or update only affected cache entries. Keep cache invalidation out of
  generic display components.
- Keep pagination and shareable filters in URL search parameters; keep temporary dialog,
  selection, and unsaved input state local. Put state in context only when multiple distant
  parts of the application genuinely need it.

### Infrastructure and API adapters

- Make HTTP calls through the shared Axios client in `lib/axios.ts`; it owns shared
  configuration, credentials, token refresh, and unauthorized-session handling.
- Keep endpoint paths, request/response shapes, and transport concerns in API adapters
  under `services/` (or a feature's `infrastructure/` as it is introduced). Components
  must not call Axios directly.
- Name API types as transport DTOs when they differ from domain concepts. Add explicit
  mapping at the boundary when API casing, nullability, or response structure should not
  leak into product logic. Avoid mapping boilerplate when the shapes are identical and
  the mapping adds no protection.
- Normalize errors at a shared boundary where possible. Pages choose user-facing
  feedback; adapters should not show toasts or depend on UI components.

### Routes, pages, and components

- Register paths in `app-routes.tsx`; use lazy loading and route-level permission guards
  consistently with existing routes.
- Keep route wrappers focused on URL parameters, access/availability checks, and choosing
  the page. Keep page components responsible for composing the screen, not for owning
  HTTP protocol details.
- Use feature components for feature-specific interactions and `components/` for UI that
  is truly reusable across features. Generic components receive data and callbacks via
  props; they do not know API endpoints or the active business module.
- Keep forms responsible for input, validation presentation, and submission intent.
  Reuse schemas where appropriate, and keep business invariants in domain/application
  logic rather than only in form validation.
- Every async screen should deliberately cover loading, failure, empty, and success
  states. Do not show sample or fabricated records as real product data.
- Preserve keyboard accessibility, semantic HTML, responsive layouts, and existing
  design-system components/tokens.

### Localization and configuration

- User-visible strings belong in `i18n/locales/` and are accessed through `useI18n()`;
  avoid hard-coded copy in product screens.
- Read environment values through `config/` and validate them centrally. Do not access
  `import.meta.env` throughout feature code.
- Keep application bootstrap concerns in `providers.tsx`; feature code should consume
  providers through their public hooks rather than reaching into provider internals.

## Testing criteria

Choose tests at the boundary where a behavior lives:

1. **Domain:** unit-test policies, role decisions, state transitions, and edge cases as
   pure functions.
2. **Application:** test query/mutation orchestration, organization scoping, and cache
   effects without coupling assertions to component markup.
3. **Presentation:** use React Testing Library for user-visible behavior—permissions,
   validation, loading/error/empty states, and actions.
4. **Infrastructure:** test request parameters and response mapping when the adapter has
   meaningful transport logic.

Prefer assertions on outcomes and user interactions over implementation details. A UI
test is not a substitute for server authorization tests.

## Review checklist for a frontend change

- Does each file have one clear responsibility and depend only on the layer below it?
- Are product rules independent from React and HTTP?
- Are organization scope and role-dependent actions explicit, with server enforcement
  still authoritative?
- Are query keys complete and are mutation cache effects correct?
- Are loading, error, empty, and success states handled?
- Are user-facing strings localized and components accessible?
- Are focused tests included for changed behavior?
- Is this a real Learner feature, rather than starter-template/demo functionality?

## Commands

Run the frontend checks in the development container:

```bash
docker compose -f compose.dev.yml exec frontend pnpm build
docker compose -f compose.dev.yml exec frontend pnpm lint
docker compose -f compose.dev.yml exec frontend pnpm test
```
