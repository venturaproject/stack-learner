import { useLocation, useNavigate } from 'react-router-dom'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { pathFor } from '@/lib/app-routes'
import { useI18n } from '@/i18n/context'

/**
 * Shared sub-navigation for the Access Control area. The sidebar only links to
 * Users — Roles and Permissions are reached from here instead, as tabs within
 * the same section, rather than as separate top-level nav entries.
 */
export function AccessControlTabs() {
  const { t } = useI18n()
  const navigate = useNavigate()
  const { pathname } = useLocation()

  const current = pathname.startsWith('/admin/roles')
    ? 'roles'
    : pathname.startsWith('/admin/permissions')
      ? 'permissions'
      : 'users'

  const handleChange = (value: string) => {
    if (value === current) return
    navigate(pathFor(`admin.${value}.index`))
  }

  return (
    <Tabs value={current} onValueChange={handleChange}>
      <TabsList>
        <TabsTrigger value="users">{t('users') || 'Usuarios'}</TabsTrigger>
        <TabsTrigger value="roles">{t('roles') || 'Roles'}</TabsTrigger>
        <TabsTrigger value="permissions">{t('permissions') || 'Permisos'}</TabsTrigger>
      </TabsList>
    </Tabs>
  )
}
