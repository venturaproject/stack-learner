import { type ReactNode } from 'react'
import { usePermission } from '@/hooks/use-permission'
import { Forbidden } from '@/components/errors/forbidden'

/**
 * Route-level RBAC guard. Renders `children` only if the current user holds the
 * given permission (or a full-access role); otherwise shows the 403 page.
 *
 * The backend enforces the same permissions on every endpoint — this guard just
 * keeps users out of pages they can't use.
 */
export function RequirePermission({
  permission,
  children,
}: {
  permission: string | string[]
  children: ReactNode
}) {
  const { checkPermission } = usePermission()
  if (!checkPermission(permission)) return <Forbidden />
  return <>{children}</>
}
