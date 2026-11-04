import { useMemo, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { CalendarDays, GraduationCap, Plus, Users, X } from 'lucide-react'
import { toast } from 'sonner'
import { AuthenticatedLayout } from '@/layouts'
import { Main } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useI18n } from '@/i18n/context'
import { learningApi, type Cohort } from '@/services/learning-api'
import { canManageOrganization } from '@/lib/learning-permissions'

function dateInput(value?: string | null) {
  if (!value) return ''
  const date = new Date(value)
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000)
  return local.toISOString().slice(0, 16)
}

function toISO(value: string) {
  return value ? new Date(value).toISOString() : null
}

export default function CohortsPage() {
  const { t } = useI18n()
  const [params, setParams] = useSearchParams()
  const queryClient = useQueryClient()
  const organizationsQuery = useQuery({ queryKey: ['learning', 'organizations'], queryFn: learningApi.organizations })
  const organizations = (organizationsQuery.data ?? []).filter((item) => !item.archived_at)
  const [organizationID, setOrganizationID] = useState<number>(() => Number(params.get('organization_id')) || 0)
  const organization = organizations.find((item) => item.id === organizationID) ?? organizations[0]
  const orgID = organization?.id ?? 0
  const coursesQuery = useQuery({ queryKey: ['learning', 'courses', orgID, 'cohorts'], queryFn: () => learningApi.courses(orgID), enabled: orgID > 0 })
  const courses = coursesQuery.data ?? []
  const [courseID, setCourseID] = useState<number>(Number(params.get('course_id')) || 0)
  const activeCourseID = courses.some((item) => item.id === courseID) ? courseID : courses[0]?.id ?? 0
  const cohortsQuery = useQuery({ queryKey: ['learning', 'cohorts', orgID, activeCourseID], queryFn: () => learningApi.cohorts(orgID, activeCourseID || undefined), enabled: orgID > 0 })
  const cohorts = cohortsQuery.data ?? []
  const [selectedID, setSelectedID] = useState<number | null>(null)
  const selected = cohorts.find((item) => item.id === selectedID) ?? cohorts[0]
  const [name, setName] = useState('')
  const [startsAt, setStartsAt] = useState('')
  const [endsAt, setEndsAt] = useState('')
  const [editingID, setEditingID] = useState<number | null>(null)
  const [studentID, setStudentID] = useState<number | null>(null)
  const [trainerID, setTrainerID] = useState<number | null>(null)
  const isAdmin = canManageOrganization(organization?.role)
  const isManager = organization?.role === 'admin' || organization?.role === 'trainer'
  const enrollmentsQuery = useQuery({
    queryKey: ['learning', 'cohort-enrollments', orgID, selected?.id],
    queryFn: () => learningApi.cohortEnrollments(orgID, selected!.id),
    enabled: orgID > 0 && selected !== undefined && isManager,
  })
  const instructorsQuery = useQuery({
    queryKey: ['learning', 'cohort-instructors', orgID, selected?.id],
    queryFn: () => learningApi.cohortInstructors(orgID, selected!.id),
    enabled: orgID > 0 && selected !== undefined && isManager,
  })
  const studentsQuery = useQuery({
    queryKey: ['learning', 'students', orgID, 'cohort-enroll'],
    queryFn: () => learningApi.studentsPage(orgID, 1, 100, ''),
    enabled: orgID > 0 && isManager,
  })
  const refresh = async () => queryClient.invalidateQueries({ queryKey: ['learning'] })
  const saveCohort = useMutation({
    mutationFn: () => {
      const input = { name: name.trim(), starts_at: toISO(startsAt), ends_at: toISO(endsAt) }
      return editingID ? learningApi.updateCohort(orgID, editingID, input) : learningApi.createCohort(orgID, activeCourseID, input)
    },
    onSuccess: async (item) => { setName(''); setStartsAt(''); setEndsAt(''); setEditingID(null); setSelectedID(item.id); await refresh(); toast.success(t(editingID ? 'cohort_updated' : 'cohort_created')) },
    onError: () => toast.error(t('cohort_save_error')),
  })
  const transition = useMutation({
    mutationFn: (status: Cohort['status']) => learningApi.transitionCohort(orgID, selected!.id, status),
    onSuccess: refresh,
    onError: () => toast.error(t('cohort_status_error')),
  })
  const enroll = useMutation({
    mutationFn: (id: number) => learningApi.enrollStudent(orgID, selected!.id, id),
    onSuccess: async () => { setStudentID(null); await refresh(); toast.success(t('cohort_student_enrolled')) },
    onError: () => toast.error(t('cohort_enroll_error')),
  })
  const withdraw = useMutation({
    mutationFn: ({ id, reason }: { id: number; reason: string }) => learningApi.withdrawStudent(orgID, selected!.id, id, reason),
    onSuccess: async () => { await refresh(); toast.success(t('cohort_student_withdrawn')) },
    onError: () => toast.error(t('cohort_withdraw_error')),
  })
  const assignTrainer = useMutation({
    mutationFn: (id: number) => learningApi.assignCohortInstructor(orgID, selected!.id, id),
    onSuccess: async () => { setTrainerID(null); await refresh(); toast.success(t('cohort_instructor_assigned')) },
    onError: () => toast.error(t('cohort_instructor_error')),
  })
  const removeTrainer = useMutation({
    mutationFn: (id: number) => learningApi.removeCohortInstructor(orgID, selected!.id, id),
    onSuccess: refresh,
    onError: () => toast.error(t('cohort_instructor_error')),
  })
  const assignedStudentIDs = useMemo(() => new Set(enrollmentsQuery.data?.filter((row) => row.status === 'active').map((row) => row.user_id)), [enrollmentsQuery.data])

  if (organizationsQuery.isLoading) return <div className="p-8 text-sm text-muted-foreground">{t('common_loading')}</div>
  if (organizationsQuery.isError) return <div className="p-8 text-sm text-destructive">{t('organizations_load_error')}</div>

  return <AuthenticatedLayout title={t('cohorts_title')}><Main>
    <div className="mx-auto w-full max-w-6xl space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div><p className="text-sm font-medium text-primary">{t('learning_workspace')}</p><h1 className="mt-2 text-3xl font-semibold">{t('cohorts_title')}</h1><p className="mt-2 text-muted-foreground">{t('cohorts_description')}</p></div>
        {organization && <label className="grid gap-1 text-sm font-medium">{t('learning_organization_label')}<select className="h-10 min-w-60 rounded-md border bg-background px-3" value={orgID} onChange={(event) => { const id = Number(event.target.value); setOrganizationID(id); setSelectedID(null); setParams({ organization_id: String(id) }) }}>{organizations.map((item) => <option key={item.id} value={item.id}>{item.name}</option>)}</select></label>}
      </header>
      {!organization ? <div className="rounded-xl border border-dashed p-10 text-center text-muted-foreground">{t('learning_no_organizations')}</div> : coursesQuery.isLoading ? <p className="text-sm text-muted-foreground">{t('learning_courses_loading')}</p> : coursesQuery.isError ? <p className="text-sm text-destructive">{t('learning_courses_load_error')}</p> : !courses.length ? <div className="rounded-xl border border-dashed p-10 text-center text-muted-foreground">{t('cohorts_no_courses')}</div> : <>
        <label className="grid max-w-xl gap-2 text-sm font-medium">{t('learning_course_name')}<select className="h-10 rounded-md border bg-background px-3" value={activeCourseID} onChange={(event) => { const id = Number(event.target.value); setCourseID(id); setSelectedID(null); setParams({ organization_id: String(orgID), course_id: String(id) }) }}>{courses.map((course) => <option key={course.id} value={course.id}>{course.title} — {t(course.status === 'published' ? 'learning_published' : 'learning_draft')}</option>)}</select></label>
        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(320px,0.8fr)]">
          <section className="space-y-4">
            {isAdmin && courses.find((course) => course.id === activeCourseID)?.status === 'published' && <form className="rounded-xl border bg-card p-5" onSubmit={(event) => { event.preventDefault(); saveCohort.mutate() }}>
              <h2 className="mb-4 font-semibold">{t(editingID ? 'cohort_edit' : 'cohort_create')}</h2>
              <div className="grid gap-3 sm:grid-cols-2"><Input required maxLength={160} placeholder={t('cohort_name')} value={name} onChange={(event) => setName(event.target.value)} /><Input aria-label={t('cohort_starts_at')} type="datetime-local" value={startsAt} onChange={(event) => setStartsAt(event.target.value)} /><Input aria-label={t('cohort_ends_at')} type="datetime-local" value={endsAt} onChange={(event) => setEndsAt(event.target.value)} /></div>
              <div className="mt-4 flex justify-end gap-2">{editingID && <Button type="button" variant="outline" onClick={() => { setEditingID(null); setName(''); setStartsAt(''); setEndsAt('') }}>{t('cancel')}</Button>}<Button disabled={saveCohort.isPending || !name.trim()}><Plus className="mr-2 h-4 w-4" />{t(editingID ? 'save' : 'cohort_create')}</Button></div>
            </form>}
            {cohortsQuery.isLoading ? <p className="text-sm text-muted-foreground">{t('common_loading')}</p> : cohortsQuery.isError ? <p className="text-sm text-destructive">{t('cohorts_load_error')}</p> : !cohorts.length ? <div className="rounded-xl border border-dashed p-8 text-center text-muted-foreground">{t('cohorts_empty')}</div> : <div className="space-y-3">{cohorts.map((cohort) => <button key={cohort.id} type="button" onClick={() => setSelectedID(cohort.id)} className={`w-full rounded-xl border bg-card p-4 text-left ${selected?.id === cohort.id ? 'border-primary ring-1 ring-primary' : ''}`}>
              <div className="flex items-start justify-between gap-3"><div><h2 className="font-semibold">{cohort.name}</h2><p className="text-sm text-muted-foreground">{cohort.course_title}</p></div><span className="rounded-full bg-muted px-2.5 py-1 text-xs">{t(`cohort_status_${cohort.status}`)}</span></div>
              <div className="mt-3 flex flex-wrap gap-4 text-xs text-muted-foreground"><span className="inline-flex items-center gap-1"><CalendarDays className="size-3.5" />{cohort.starts_at ? new Date(cohort.starts_at).toLocaleString() : t('cohort_no_start_date')}</span><span className="inline-flex items-center gap-1"><Users className="size-3.5" />{cohort.enrollment_count} {t('cohort_students_count')}</span></div>
            </button>)}</div>}
          </section>
          {selected && <aside className="space-y-4 rounded-xl border bg-card p-5">
            <div className="flex items-start justify-between gap-3"><div><p className="text-sm text-muted-foreground">{selected.course_title}</p><h2 className="text-xl font-semibold">{selected.name}</h2><p className="mt-1 text-sm">{t(`cohort_status_${selected.status}`)}</p></div><GraduationCap className="size-6 text-primary" /></div>
            {isAdmin && <div className="flex flex-wrap gap-2">{selected.status === 'draft' && <><Button size="sm" onClick={() => transition.mutate('active')}>{t('cohort_activate')}</Button><Button size="sm" variant="outline" onClick={() => { setEditingID(selected.id); setName(selected.name); setStartsAt(dateInput(selected.starts_at)); setEndsAt(dateInput(selected.ends_at)) }}>{t('edit')}</Button></>}{selected.status === 'active' && <><Button size="sm" onClick={() => transition.mutate('completed')}>{t('cohort_complete')}</Button><Button size="sm" variant="outline" onClick={() => transition.mutate('cancelled')}>{t('cohort_cancel')}</Button></>}</div>}
            {(isAdmin || (organization?.role === 'trainer')) && <>
              {selected.status === 'active' && <div className="space-y-2 border-t pt-4"><p className="text-sm font-medium">{t('cohort_enroll_student')}</p><div className="flex gap-2"><select aria-label={t('cohort_select_student')} className="h-9 min-w-0 flex-1 rounded-md border bg-background px-2 text-sm" value={studentID ?? ''} onChange={(event) => setStudentID(Number(event.target.value) || null)}><option value="">{t('cohort_select_student')}</option>{studentsQuery.data?.items.filter((student) => !assignedStudentIDs.has(student.user_id)).map((student) => <option key={student.user_id} value={student.user_id}>{student.name} ({student.email})</option>)}</select><Button size="sm" disabled={studentID === null || enroll.isPending} onClick={() => { if (studentID !== null) enroll.mutate(studentID) }}>{t('cohort_enroll')}</Button></div></div>}
              {isAdmin && <div className="space-y-2 border-t pt-4"><p className="text-sm font-medium">{t('cohort_instructors')}</p><div className="flex flex-wrap gap-2">{instructorsQuery.data?.filter((item) => item.assigned).map((item) => <span key={item.user_id} className="inline-flex items-center gap-1 rounded-full bg-muted px-3 py-1 text-xs">{item.name}<button type="button" aria-label={t('cohort_remove_instructor', { name: item.name })} onClick={() => removeTrainer.mutate(item.user_id)}><X className="size-3" /></button></span>)}</div><div className="flex gap-2"><select aria-label={t('cohort_select_instructor')} className="h-9 min-w-0 flex-1 rounded-md border bg-background px-2 text-sm" value={trainerID ?? ''} onChange={(event) => setTrainerID(Number(event.target.value) || null)}><option value="">{t('cohort_select_instructor')}</option>{instructorsQuery.data?.filter((item) => !item.assigned).map((item) => <option key={item.user_id} value={item.user_id}>{item.name} ({item.email})</option>)}</select><Button size="sm" disabled={trainerID === null} onClick={() => { if (trainerID !== null) assignTrainer.mutate(trainerID) }}>{t('cohort_assign_instructor')}</Button></div></div>}
              <div className="space-y-2 border-t pt-4"><h3 className="text-sm font-medium">{t('cohort_enrollments')}</h3>{enrollmentsQuery.isLoading ? <p className="text-sm text-muted-foreground">{t('common_loading')}</p> : enrollmentsQuery.isError ? <p className="text-sm text-destructive">{t('cohort_enrollments_error')}</p> : enrollmentsQuery.data?.length ? <ul className="divide-y">{enrollmentsQuery.data.map((enrollment) => <li key={enrollment.id} className="flex items-center justify-between gap-2 py-2 text-sm"><span>{enrollment.name}<span className="ml-2 text-xs text-muted-foreground">{enrollment.email} · {t(`cohort_enrollment_${enrollment.status}`)}</span></span>{enrollment.status === 'active' && <Button type="button" size="sm" variant="ghost" className="text-destructive" onClick={() => { const reason = window.prompt(t('cohort_withdraw_reason')); if (reason !== null) withdraw.mutate({ id: enrollment.user_id, reason }) }}>{t('cohort_withdraw')}</Button>}</li>)}</ul> : <p className="text-sm text-muted-foreground">{t('cohort_enrollments_empty')}</p>}</div>
            </>}
          </aside>}
        </div>
      </>}
    </div>
  </Main></AuthenticatedLayout>
}
