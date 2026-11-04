/**
 * Application route registry.
 *
 * This is the only file that should need to change when adding, removing,
 * or renaming business-module routes. The core router (lib/router.tsx) has
 * no knowledge of these paths.
 */
import { lazy, Suspense, type ReactNode } from 'react'
import type { RouteObject } from 'react-router-dom'
import { RoutePending } from '@/components/route-pending'

// Auth pages
const SignIn = lazy(() => import('@/pages/auth/sign-in/sign-in-2'))
const ForgotPasswordPage = lazy(() => import('@/pages/auth/forgot-password'))

// Route wrappers (each wraps one or more pages for a domain module)
const DashboardRoute = lazy(() => import('@/routes/dashboard-route'))
const LearningRoute = lazy(() => import('@/routes/learning-route'))
const CohortsPage = lazy(() => import('@/pages/cohorts'))
const OrganizationsRoute = lazy(() => import('@/routes/organizations-route'))
const StudentsRoute = lazy(() => import('@/pages/students'))
const CreateStudentPage = lazy(() => import('@/pages/students/create'))
const UsersRoute = lazy(() => import('@/routes/users-route'))
const RolesRoute = lazy(() => import('@/routes/roles-route'))
const PermissionsRoute = lazy(() => import('@/routes/permissions-route'))
const TelefonosRoute = lazy(() => import('@/routes/telefonos-route'))
const DispositivosRoute = lazy(() => import('@/routes/dispositivos-route'))
const TrabajadoresRoute = lazy(() => import('@/routes/trabajadores-route'))
const SettingsRoute = lazy(() => import('@/routes/settings-route'))
import { RequirePermission } from '@/components/require-permission'

const load = (Page: React.ComponentType) => (
  <Suspense fallback={<RoutePending />}>
    <Page />
  </Suspense>
)

// ── Guest routes (unauthenticated only) ───────────────────────────────────────

export const guestRoutes: RouteObject[] = [
  { path: '/login', element: <Suspense fallback={<RoutePending />}><SignIn canResetPassword /></Suspense> },
  { path: '/forgot-password', element: load(ForgotPasswordPage) },
]

// ── Private routes (authenticated only) ──────────────────────────────────────

const guard = (permission: string | string[], element: ReactNode) => (
  <RequirePermission permission={permission}>{element}</RequirePermission>
)

export const privateRoutes: RouteObject[] = [
  // Dashboard is the post-login landing page; LMS sections are separate routes.
  { path: '/admin', element: load(DashboardRoute) },
  { path: '/admin/learning', element: load(LearningRoute) },
  { path: '/admin/cohorts', element: load(CohortsPage) },
  { path: '/admin/learning/courses/create', element: load(LearningRoute) },
  { path: '/admin/learning/courses/:id/edit', element: load(LearningRoute) },
  { path: '/admin/organizations', element: load(OrganizationsRoute) },
  { path: '/admin/organizations/create', element: load(OrganizationsRoute) },
  { path: '/admin/organizations/:id/edit', element: load(OrganizationsRoute) },
  { path: '/admin/students', element: load(StudentsRoute) },
  { path: '/admin/students/create', element: load(CreateStudentPage) },

  // Users module
  { path: '/admin/users', element: guard('users.view', load(UsersRoute)) },
  { path: '/admin/users/create', element: guard('users.create', load(UsersRoute)) },
  { path: '/admin/users/:id', element: guard('users.view', load(UsersRoute)) },
  { path: '/admin/users/:id/edit', element: guard('users.edit', load(UsersRoute)) },

  // Roles module
  { path: '/admin/roles', element: guard('roles.view', load(RolesRoute)) },
  { path: '/admin/roles/create', element: guard('roles.create', load(RolesRoute)) },
  { path: '/admin/roles/:id/edit', element: guard('roles.edit', load(RolesRoute)) },

  // Permissions module
  { path: '/admin/permissions', element: guard('permissions.view', load(PermissionsRoute)) },
  { path: '/admin/permissions/create', element: guard('permissions.create', load(PermissionsRoute)) },
  { path: '/admin/permissions/:id/edit', element: guard('permissions.view', load(PermissionsRoute)) },

  // Telefonos module
  { path: '/admin/telefonos', element: guard('telefonos.view', load(TelefonosRoute)) },
  { path: '/admin/telefonos/:id', element: guard('telefonos.view', load(TelefonosRoute)) },
  { path: '/admin/telefonos/:id/edit', element: guard('telefonos.edit', load(TelefonosRoute)) },

  // Dispositivos module
  { path: '/admin/dispositivos', element: guard('dispositivos.view', load(DispositivosRoute)) },
  { path: '/admin/dispositivos/create', element: guard('dispositivos.create', load(DispositivosRoute)) },
  { path: '/admin/dispositivos/:id', element: guard('dispositivos.view', load(DispositivosRoute)) },
  { path: '/admin/dispositivos/:id/edit', element: guard('dispositivos.edit', load(DispositivosRoute)) },

  // Trabajadores module
  { path: '/admin/trabajadores', element: guard('trabajadores.view', load(TrabajadoresRoute)) },
  { path: '/admin/trabajadores/:id', element: guard('trabajadores.view', load(TrabajadoresRoute)) },

  // Settings
  { path: '/admin/settings', element: load(SettingsRoute) },
  { path: '/admin/settings/:section', element: load(SettingsRoute) },
]
