import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { ColumnDef } from '@tanstack/react-table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Archive, MoreHorizontal, Pencil, Plus, RotateCcw } from 'lucide-react'
import { toast } from 'sonner'
import { AuthenticatedLayout } from '@/layouts'
import { Main } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { useI18n } from '@/i18n/context'
import { learningApi, type Organization } from '@/services/learning-api'
import { DataTablePanel } from '@/components/data-table'
import { canManageOrganization } from '@/lib/learning-permissions'

export default function OrganizationsPage() {
  const { t } = useI18n()
  const navigate = useNavigate()
  const client = useQueryClient()
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(10)
  const [search, setSearch] = useState('')
  const organizations = useQuery({
    queryKey: ['learning', 'organizations', page, perPage, search],
    queryFn: () => learningApi.organizationsPage(page, perPage, search),
  })
  const [deleteTarget, setDeleteTarget] = useState<Organization | null>(null)
  const refresh = () => client.invalidateQueries({ queryKey: ['learning', 'organizations'] })
  const remove = useMutation({
    mutationFn: learningApi.deleteOrganization,
    onSuccess: async () => { await refresh(); toast.success(t('organization_deleted')) },
    onError: () => toast.error(t('organization_delete_error')),
  })
  const restore = useMutation({
    mutationFn: learningApi.restoreOrganization,
    onSuccess: async () => { await refresh(); toast.success(t('organization_restored')) },
    onError: () => toast.error(t('organization_restore_error')),
  })
  const restoreOrganization = restore.mutate
  const columns = useMemo<ColumnDef<Organization>[]>(() => [
    { accessorKey: 'name', header: t('organization_name'), cell: ({ row }) => <span className="font-medium">{row.original.name}</span> },
    { accessorKey: 'slug', header: t('organization_slug'), cell: ({ row }) => <span className="text-muted-foreground">{row.original.slug}</span> },
    { id: 'actions', enableHiding: false, enableSorting: false, header: () => <span className="sr-only">{t('actions')}</span>, cell: ({ row }) => (
      <DropdownMenu modal={false}><DropdownMenuTrigger asChild><Button variant="ghost" size="icon" aria-label={t('actions')}><MoreHorizontal className="h-4 w-4" /></Button></DropdownMenuTrigger><DropdownMenuContent align="end">
        {row.original.archived_at ? <DropdownMenuItem disabled={!canManageOrganization(row.original.role)} onClick={() => restoreOrganization(row.original.id)}><RotateCcw className="mr-2 h-4 w-4" />{t('organization_restore')}</DropdownMenuItem> : canManageOrganization(row.original.role) && <><DropdownMenuItem onClick={() => navigate(`/admin/organizations/${row.original.id}/edit`)}><Pencil className="mr-2 h-4 w-4" />{t('edit')}</DropdownMenuItem><DropdownMenuSeparator /><DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => setDeleteTarget(row.original)}><Archive className="mr-2 h-4 w-4" />{t('organization_archive')}</DropdownMenuItem></>}
      </DropdownMenuContent></DropdownMenu>
    ) },
  ], [navigate, restoreOrganization, t])

  return (
    <AuthenticatedLayout title={t('learning_organizations_navigation')}>
      <Main>
        <div className="grid flex-1 items-start gap-4 md:gap-8">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"><div><h1 className="text-2xl font-bold tracking-tight">{t('learning_organizations_navigation')}</h1><p className="text-muted-foreground">{t('organizations_description')}</p></div><Button onClick={() => navigate('/admin/organizations/create')} className="w-fit"><Plus className="mr-2 h-4 w-4" />{t('organization_add')}</Button></div>

          {organizations.isLoading ? <p className="text-muted-foreground">{t('common_loading')}</p> : organizations.isError ? <p className="text-destructive">{t('organizations_load_error')}</p> : (
            <DataTablePanel data={organizations.data?.items ?? []} columns={columns} title={t('all_organizations_title')} description={t('all_organizations_description')} emptyMessage={t('organizations_empty')} searchPlaceholder={t('organizations_search')} columnLabels={{ name: t('organization_name'), slug: t('organization_slug') }} serverPage={{ page, perPage, total: organizations.data?.pagination.total ?? 0, totalPages: organizations.data?.pagination.total_pages ?? 1, onPageChange: setPage, onPerPageChange: (size) => { setPerPage(size); setPage(1) }, onSearchChange: (value) => { setSearch(value); setPage(1) } }} />
          )}
        </div>
      </Main>
      <ConfirmDialog open={deleteTarget !== null} onOpenChange={(open) => { if (!open) setDeleteTarget(null) }} title={t('organization_archive')} desc={t('organization_archive_confirm', { name: deleteTarget?.name ?? '' })} confirmText={t('organization_archive')} destructive handleConfirm={() => { if (deleteTarget) remove.mutate(deleteTarget.id); setDeleteTarget(null) }} />
    </AuthenticatedLayout>
  )
}
