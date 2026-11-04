import { useMemo, useState } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { GraduationCap, MoreHorizontal, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { AuthenticatedLayout } from '@/layouts'
import { Main } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/confirm-dialog'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { useI18n } from '@/i18n/context'
import { learningApi, type Student } from '@/services/learning-api'
import { DataTablePanel } from '@/components/data-table'
import { canManageOrganization } from '@/lib/learning-permissions'

export default function StudentsPage() {
  const { t } = useI18n()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const client = useQueryClient()
  const organizations = useQuery({ queryKey: ['learning', 'organizations'], queryFn: learningApi.organizations })
  const activeOrganizations = (organizations.data ?? []).filter((organization) => !organization.archived_at)
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(10)
  const [search, setSearch] = useState('')
  const [organizationID, setOrganizationID] = useState<number | null>(() => Number(searchParams.get('organization_id')) || null)
  const [removeTarget, setRemoveTarget] = useState<Student | null>(null)
  const activeOrganizationID = organizationID && activeOrganizations.some((organization) => organization.id === organizationID)
    ? organizationID
    : activeOrganizations[0]?.id ?? null
  const canManageMemberships = canManageOrganization(activeOrganizations.find((organization) => organization.id === activeOrganizationID)?.role)
  const students = useQuery({
    queryKey: ['learning', 'students', activeOrganizationID, page, perPage, search],
    queryFn: () => learningApi.studentsPage(activeOrganizationID!, page, perPage, search),
    enabled: activeOrganizationID !== null,
  })
  const refresh = () => client.invalidateQueries({ queryKey: ['learning', 'students', activeOrganizationID] })
  const remove = useMutation({
    mutationFn: (studentID: number) => learningApi.deleteStudent(activeOrganizationID!, studentID),
    onSuccess: async () => { await refresh(); toast.success(t('student_removed')) },
    onError: () => toast.error(t('student_remove_error')),
  })
  const columns = useMemo<ColumnDef<Student>[]>(() => [
    { accessorKey: 'name', header: t('student_name'), cell: ({ row }) => <span className="font-medium">{row.original.name}</span> },
    { accessorKey: 'email', header: t('student_email'), cell: ({ row }) => <span className="text-muted-foreground">{row.original.email}</span> },
    { id: 'actions', enableHiding: false, enableSorting: false, header: () => <span className="sr-only">{t('actions')}</span>, cell: ({ row }) => (
      <DropdownMenu modal={false}><DropdownMenuTrigger asChild><Button variant="ghost" size="icon" aria-label={t('actions')}><MoreHorizontal className="h-4 w-4" /></Button></DropdownMenuTrigger><DropdownMenuContent align="end">
        {canManageMemberships && <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => setRemoveTarget(row.original)}><Trash2 className="mr-2 h-4 w-4" />{t('student_remove')}</DropdownMenuItem>}
      </DropdownMenuContent></DropdownMenu>
    ) },
  ], [canManageMemberships, t])

  return (
    <AuthenticatedLayout title={t('learning_students_navigation')}>
      <Main>
        <div className="grid flex-1 items-start gap-4 md:gap-8">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"><div><h1 className="text-2xl font-bold tracking-tight">{t('learning_students_navigation')}</h1><p className="text-muted-foreground">{t('students_description')}</p></div>{activeOrganizationID && canManageMemberships && <Button onClick={() => navigate(`/admin/students/create?organization_id=${activeOrganizationID}`)} className="w-fit"><Plus className="mr-2 h-4 w-4" />{t('student_add')}</Button>}</div>

          {organizations.isLoading ? <p className="text-muted-foreground">{t('common_loading')}</p> : organizations.isError ? <p className="text-destructive">{t('organizations_load_error')}</p> : !activeOrganizations.length ? (
            <div className="rounded-xl border border-dashed p-10 text-center text-muted-foreground"><GraduationCap className="mx-auto mb-3" />{t('students_need_organization')}</div>
          ) : (
            <>
              <label className="grid max-w-sm gap-2 text-sm font-medium">
                {t('learning_organization_label')}
                  <select className="h-10 rounded-md border bg-background px-3" value={activeOrganizationID ?? ''} onChange={(event) => { const id = Number(event.target.value); setOrganizationID(id); setPage(1); setSearchParams({ organization_id: String(id) }) }}>
                  {activeOrganizations.map((organization) => <option key={organization.id} value={organization.id}>{organization.name}</option>)}
                </select>
              </label>

              {students.isLoading ? <p className="text-muted-foreground">{t('common_loading')}</p> : students.isError ? <p className="text-destructive">{t('students_load_error')}</p> : (
                 <DataTablePanel data={students.data?.items ?? []} columns={columns} title={t('all_students_title')} description={t('all_students_description')} emptyMessage={t('students_empty')} searchPlaceholder={t('students_search')} columnLabels={{ name: t('student_name'), email: t('student_email') }} serverPage={{ page, perPage, total: students.data?.pagination.total ?? 0, totalPages: students.data?.pagination.total_pages ?? 1, onPageChange: setPage, onPerPageChange: (size) => { setPerPage(size); setPage(1) }, onSearchChange: (value) => { setSearch(value); setPage(1) } }} />
              )}
            </>
          )}
        </div>
      </Main>
      <ConfirmDialog open={removeTarget !== null} onOpenChange={(open) => { if (!open) setRemoveTarget(null) }} title={t('student_remove')} desc={t('student_remove_confirm', { name: removeTarget?.name ?? '' })} confirmText={t('student_remove')} destructive handleConfirm={() => { if (removeTarget) remove.mutate(removeTarget.user_id); setRemoveTarget(null) }} />
    </AuthenticatedLayout>
  )
}
