import { useMemo, useState, type ReactNode } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { ColumnDef } from '@tanstack/react-table'
import { ArrowDown, ArrowUp, BookOpen, Building2, CheckCircle2, GraduationCap, Layers3, LockKeyhole, MoreHorizontal, Pencil, Plus, Send, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { AuthenticatedLayout } from '@/layouts'
import { Main } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { useI18n } from '@/i18n/context'
import { learningApi, type AssessmentSubmission, type Course } from '@/services/learning-api'
import { DataTablePanel } from '@/components/data-table'
import { canAssignCourseInstructors, canManageCourse } from '@/lib/learning-permissions'

export default function LearningPage() {
  const { t } = useI18n()
  const client = useQueryClient()
  const navigate = useNavigate()
  const [searchParams, setSearchParams] = useSearchParams()
  const [page, setPage] = useState(1)
  const [perPage, setPerPage] = useState(10)
  const [courseSearch, setCourseSearch] = useState('')
  const organizationsQuery = useQuery({ queryKey: ['learning', 'organizations'], queryFn: learningApi.organizations })
  const organizations = (organizationsQuery.data ?? []).filter((organization) => !organization.archived_at)
  const [organizationID, setOrganizationID] = useState<number | null>(() => Number(searchParams.get('organization_id')) || null)
  const [selectedCourseID, setSelectedCourseID] = useState<number | null>(null)
  const [moduleTitle, setModuleTitle] = useState('')
  const [lessonTitle, setLessonTitle] = useState('')
  const [lessonKind, setLessonKind] = useState<'text' | 'task'>('text')
  const [lessonModuleID, setLessonModuleID] = useState<number | null>(null)
  const [trainerToAssign, setTrainerToAssign] = useState<number | null>(null)
  const activeOrganizationID = organizationID && organizations.some((organization) => organization.id === organizationID)
    ? organizationID
    : organizations[0]?.id ?? null
  const activeOrganization = organizations.find((organization) => organization.id === activeOrganizationID)
  const canManage = canManageCourse(activeOrganization?.role)
  const coursesQuery = useQuery({
    queryKey: ['learning', 'courses', activeOrganizationID, page, perPage, courseSearch],
    queryFn: () => learningApi.coursesPage(activeOrganizationID!, page, perPage, courseSearch),
    enabled: activeOrganizationID !== null,
  })
  const courseQuery = useQuery({
    queryKey: ['learning', 'course', activeOrganizationID, selectedCourseID],
    queryFn: () => learningApi.course(activeOrganizationID!, selectedCourseID!),
    enabled: activeOrganizationID !== null && selectedCourseID !== null,
  })
  const selected = courseQuery.data ?? coursesQuery.data?.items.find((course) => course.id === selectedCourseID)
  const refresh = async () => client.invalidateQueries({ queryKey: ['learning'] })
  const deleteCourse = useMutation({
    mutationFn: (courseID: number) => learningApi.deleteCourse(activeOrganizationID!, courseID),
    onSuccess: async () => { setSelectedCourseID(null); await refresh(); toast.success(t('learning_course_deleted')) },
    onError: () => toast.error(t('learning_course_delete_error')),
  })
  const addModule = useMutation({
    mutationFn: () => learningApi.createModule(activeOrganizationID!, selectedCourseID!, moduleTitle),
    onSuccess: async () => { setModuleTitle(''); await refresh(); toast.success(t('learning_module_added')) },
    onError: () => toast.error(t('learning_module_add_error')),
  })
  const addLesson = useMutation({
    mutationFn: ({ moduleID, title, kind }: { moduleID: number; title: string; kind: 'text' | 'task' }) => learningApi.createLesson(activeOrganizationID!, selectedCourseID!, moduleID, title, kind),
    onSuccess: async () => { setLessonTitle(''); setLessonKind('text'); setLessonModuleID(null); await refresh(); toast.success(t('learning_lesson_added')) },
    onError: () => toast.error(t('learning_lesson_add_error')),
  })
  const updateModule = useMutation({
    mutationFn: ({ moduleID, title }: { moduleID: number; title: string }) => learningApi.updateModule(activeOrganizationID!, selectedCourseID!, moduleID, title),
    onSuccess: refresh,
    onError: () => toast.error(t('learning_module_update_error')),
  })
  const deleteModule = useMutation({
    mutationFn: (moduleID: number) => learningApi.deleteModule(activeOrganizationID!, selectedCourseID!, moduleID),
    onSuccess: async () => { await refresh(); toast.success(t('learning_module_deleted')) },
    onError: () => toast.error(t('learning_module_delete_error')),
  })
  const reorderModules = useMutation({
    mutationFn: (ids: number[]) => learningApi.reorderModules(activeOrganizationID!, selectedCourseID!, ids),
    onSuccess: refresh,
    onError: () => toast.error(t('learning_reorder_error')),
  })
  const updateLesson = useMutation({
    mutationFn: ({ moduleID, lessonID, title, kind, content }: { moduleID: number; lessonID: number; title: string; kind: string; content: string }) => learningApi.updateLesson(activeOrganizationID!, selectedCourseID!, moduleID, lessonID, title, kind, content),
    onSuccess: refresh,
    onError: () => toast.error(t('learning_lesson_update_error')),
  })
  const deleteLesson = useMutation({
    mutationFn: ({ moduleID, lessonID }: { moduleID: number; lessonID: number }) => learningApi.deleteLesson(activeOrganizationID!, selectedCourseID!, moduleID, lessonID),
    onSuccess: async () => { await refresh(); toast.success(t('learning_lesson_deleted')) },
    onError: () => toast.error(t('learning_lesson_delete_error')),
  })
  const reorderLessons = useMutation({
    mutationFn: ({ moduleID, ids }: { moduleID: number; ids: number[] }) => learningApi.reorderLessons(activeOrganizationID!, selectedCourseID!, moduleID, ids),
    onSuccess: refresh,
    onError: () => toast.error(t('learning_reorder_error')),
  })
  const publishCourse = useMutation({
    mutationFn: () => learningApi.publishCourse(activeOrganizationID!, selectedCourseID!),
    onSuccess: async () => { await refresh(); toast.success(t('learning_course_published')) },
    onError: () => toast.error(t('learning_course_publish_error')),
  })
  const canAssignTrainers = canAssignCourseInstructors(activeOrganization?.role)
  const courseTrainersQuery = useQuery({
    queryKey: ['learning', 'course-trainers', activeOrganizationID, selectedCourseID],
    queryFn: () => learningApi.courseTrainers(activeOrganizationID!, selectedCourseID!),
    enabled: canManage && activeOrganizationID !== null && selectedCourseID !== null,
  })
  const isStudent = activeOrganization?.role === 'student'
  const courseProgressQuery = useQuery({
    queryKey: ['learning', 'course-progress', activeOrganizationID, selectedCourseID],
    queryFn: () => learningApi.courseProgress(activeOrganizationID!, selectedCourseID!),
    enabled: isStudent && activeOrganizationID !== null && selectedCourseID !== null,
  })
  const completeLesson = useMutation({
    mutationFn: (lessonID: number) => learningApi.completeLesson(activeOrganizationID!, selectedCourseID!, lessonID),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['learning', 'course-progress', activeOrganizationID, selectedCourseID] })
      toast.success(t('learning_lesson_completed_success'))
    },
    onError: () => toast.error(t('learning_lesson_completed_error')),
  })
  const assignTrainer = useMutation({
    mutationFn: (trainerID: number) => learningApi.assignCourseTrainer(activeOrganizationID!, selectedCourseID!, trainerID),
    onSuccess: async () => { setTrainerToAssign(null); await refresh(); toast.success(t('learning_trainer_assigned')) },
    onError: () => toast.error(t('learning_trainer_assign_error')),
  })
  const unassignTrainer = useMutation({
    mutationFn: (trainerID: number) => learningApi.removeCourseTrainer(activeOrganizationID!, selectedCourseID!, trainerID),
    onSuccess: async () => { await refresh(); toast.success(t('learning_trainer_removed')) },
    onError: () => toast.error(t('learning_trainer_remove_error')),
  })

  const courseColumns = useMemo<ColumnDef<Course>[]>(() => [
    { accessorKey: 'title', header: t('learning_course_name'), cell: ({ row }) => <Button variant="link" className="h-auto p-0 font-medium" onClick={() => setSelectedCourseID(row.original.id)}>{row.original.title}</Button> },
    { accessorKey: 'description', header: t('learning_course_description'), cell: ({ row }) => <span className="block max-w-sm truncate text-muted-foreground">{row.original.description || t('learning_no_description')}</span> },
    { accessorKey: 'status', header: t('learning_status'), cell: ({ row }) => <span className="rounded-full bg-muted px-2.5 py-1 text-xs font-medium">{t(row.original.status === 'published' ? 'learning_published' : 'learning_draft')}</span> },
    { id: 'actions', enableHiding: false, enableSorting: false, header: () => <span className="sr-only">{t('actions')}</span>, cell: ({ row }) => (
      <DropdownMenu modal={false}><DropdownMenuTrigger asChild><Button variant="ghost" size="icon" aria-label={t('actions')}><MoreHorizontal className="h-4 w-4" /></Button></DropdownMenuTrigger><DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => setSelectedCourseID(row.original.id)}><BookOpen className="mr-2 h-4 w-4" />{t('learning_open_course')}</DropdownMenuItem>
        {canManage && row.original.status === 'draft' && <DropdownMenuItem onClick={() => navigate(`/admin/learning/courses/${row.original.id}/edit?organization_id=${activeOrganizationID}`)}><Pencil className="mr-2 h-4 w-4" />{t('edit')}</DropdownMenuItem>}
        {canManage && row.original.status === 'draft' && <><DropdownMenuSeparator /><DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => { if (window.confirm(t('learning_course_delete_confirm', { title: row.original.title }))) deleteCourse.mutate(row.original.id) }}><Trash2 className="mr-2 h-4 w-4" />{t('delete')}</DropdownMenuItem></>}
      </DropdownMenuContent></DropdownMenu>
    ) },
  ], [activeOrganizationID, canManage, deleteCourse, navigate, t])

  if (organizationsQuery.isLoading) return <div className="p-8 text-sm text-muted-foreground">{t('learning_loading')}</div>
  if (organizationsQuery.isError) return <div className="p-8 text-sm text-destructive">{t('organizations_load_error')}</div>

  return (
    <AuthenticatedLayout title={t('learning_title')}>
      <Main>
        <div className="w-full space-y-6">
          <header className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
            <div>
              <p className="text-sm font-medium text-primary">{t('learning_workspace')}</p>
              <h1 className="mt-2 text-3xl font-semibold tracking-tight">{t('learning_title')}</h1>
              <p className="mt-2 max-w-2xl text-muted-foreground">{t('learning_description')}</p>
            </div>
            {organizations.length > 0 && <label className="grid gap-1 text-sm font-medium">{t('learning_organization_label')}
              <select className="h-10 min-w-60 rounded-md border bg-background px-3" value={activeOrganizationID ?? ''} onChange={(event) => { const id = Number(event.target.value); setOrganizationID(id); setSearchParams({ organization_id: String(id) }); setSelectedCourseID(null) }}>
                {organizations.map((organization) => <option key={organization.id} value={organization.id}>{organization.name}</option>)}
              </select>
            </label>}
          </header>

          {organizations.length === 0 ? (
            <section className="rounded-xl border border-dashed p-10 text-center">
              <Building2 className="mx-auto mb-3 text-muted-foreground" />
              <p className="mb-4 text-muted-foreground">{t('learning_no_organizations')}</p>
              <Link className="inline-flex h-10 items-center gap-2 rounded-md bg-primary px-4 text-primary-foreground" to="/admin/organizations"><Plus size={16} />{t('learning_create_organization')}</Link>
            </section>
          ) : (
            <>
              <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
                <StatCard
                  icon={<BookOpen />}
                  label={t('learning_courses_label')}
                  value={coursesQuery.data?.pagination.summary?.course_count ?? 0}
                  detail={`${coursesQuery.data?.pagination.summary?.published_count ?? 0} ${t('learning_published').toLowerCase()} · ${Math.max(0, (coursesQuery.data?.pagination.summary?.course_count ?? 0) - (coursesQuery.data?.pagination.summary?.published_count ?? 0))} ${t('learning_drafts').toLowerCase()}`}
                  chart={<CourseStatusChart published={coursesQuery.data?.pagination.summary?.published_count ?? 0} total={coursesQuery.data?.pagination.summary?.course_count ?? 0} />}
                />
                <StatCard
                  icon={<GraduationCap />}
                  label={t('learning_published')}
                  value={coursesQuery.data?.pagination.summary?.published_count ?? 0}
                  detail={t('learning_published_share', { percent: coursesQuery.data?.pagination.summary?.course_count ? Math.round((coursesQuery.data.pagination.summary.published_count / coursesQuery.data.pagination.summary.course_count) * 100) : 0 })}
                  chart={<ProgressRing value={coursesQuery.data?.pagination.summary?.course_count ? (coursesQuery.data.pagination.summary.published_count / coursesQuery.data.pagination.summary.course_count) * 100 : 0} />}
                />
                <StatCard
                  icon={<Layers3 />}
                  label={t('learning_modules_created')}
                  value={coursesQuery.data?.pagination.summary?.module_count ?? 0}
                  detail={t('learning_modules_per_course', { count: coursesQuery.data?.pagination.summary?.course_count ? ((coursesQuery.data.pagination.summary.module_count / coursesQuery.data.pagination.summary.course_count).toFixed(1)) : '0' })}
                  chart={<ModuleBars value={coursesQuery.data?.pagination.summary?.module_count ?? 0} />}
                />
              </section>

              <section className={`grid gap-6 ${selected ? 'lg:grid-cols-[minmax(0,1fr)_minmax(320px,0.8fr)]' : 'grid-cols-1'}`}>
                {coursesQuery.isLoading ? <p className="text-sm text-muted-foreground">{t('learning_courses_loading')}</p> : coursesQuery.isError ? <p className="text-sm text-destructive">{t('learning_courses_load_error')}</p> : (
                  <DataTablePanel data={coursesQuery.data?.items ?? []} columns={courseColumns} title={t('learning_course_catalog')} description={t('learning_course_catalog_description')} emptyMessage={t('learning_no_courses')} searchPlaceholder={t('learning_courses_search')} columnLabels={{ title: t('learning_course_name'), description: t('learning_course_description'), status: t('learning_status') }} createAction={canManage ? <Button size="sm" className="h-9 gap-1" onClick={() => navigate(`/admin/learning/courses/create?organization_id=${activeOrganizationID}`)}><Plus className="h-4 w-4" />{t('learning_new_course')}</Button> : undefined} serverPage={{ page, perPage, total: coursesQuery.data?.pagination.total ?? 0, totalPages: coursesQuery.data?.pagination.total_pages ?? 1, onPageChange: setPage, onPerPageChange: (size) => { setPerPage(size); setPage(1) }, onSearchChange: (value) => { setCourseSearch(value); setPage(1) } }} />
                )}

                <div className="space-y-5">
                  {selected && <div className="rounded-xl border bg-card p-6 shadow-sm">
                    <div className="flex items-start justify-between gap-3">
                      <div><h2 className="text-lg font-semibold">{t('learning_course_structure')}</h2><p className="text-sm text-muted-foreground">{selected.title}</p></div>
                      <div className="flex shrink-0 gap-2">
                        {canManage && selected.status === 'draft' && <button type="button" onClick={() => navigate(`/admin/learning/courses/${selected.id}/edit?organization_id=${activeOrganizationID}`)} aria-label={t('learning_edit_course')} className="inline-flex size-9 items-center justify-center rounded-md border"><Pencil size={15} /></button>}
                        {canManage && selected.status === 'draft' && <button type="button" onClick={() => { if (window.confirm(t('learning_course_delete_confirm', { title: selected.title }))) deleteCourse.mutate(selected.id) }} aria-label={t('delete')} className="inline-flex size-9 items-center justify-center rounded-md border border-destructive/30 text-destructive"><Trash2 size={15} /></button>}
                        {canManage && selected.status === 'draft' && <button type="button" onClick={() => publishCourse.mutate()} disabled={publishCourse.isPending} className="inline-flex h-9 items-center gap-2 rounded-md border px-3 text-sm"><Send size={15} />{t('learning_publish')}</button>}
                      </div>
                    </div>
                    {isStudent && <div className="mt-4 rounded-lg bg-muted/50 p-4">
                      {courseProgressQuery.isLoading ? <p className="text-sm text-muted-foreground">{t('common_loading')}</p> : courseProgressQuery.isError ? <p className="text-sm text-destructive">{t('learning_progress_load_error')}</p> : <>
                        <div className="flex items-center justify-between gap-3 text-sm"><span className="font-medium">{t('learning_progress')}</span><span>{t('learning_progress_lessons', { completed: courseProgressQuery.data?.completed_lessons ?? 0, total: courseProgressQuery.data?.total_lessons ?? 0 })}</span></div>
                        <div className="mt-2 h-2 overflow-hidden rounded-full bg-background"><div className="h-full rounded-full bg-primary transition-all" style={{ width: `${courseProgressQuery.data?.percent ?? 0}%` }} /></div>
                      </>}
                    </div>}
                    {canAssignTrainers && <div className="mt-4 space-y-2 border-t pt-4">
                      <p className="text-sm font-medium">{t('learning_course_instructors')}</p>
                      {courseTrainersQuery.isLoading ? <p className="text-sm text-muted-foreground">{t('common_loading')}</p> : courseTrainersQuery.isError ? <p className="text-sm text-destructive">{t('learning_trainers_load_error')}</p> : <>
                        <div className="flex flex-wrap gap-2">{courseTrainersQuery.data?.filter((trainer) => trainer.assigned).map((trainer) => <span key={trainer.user_id} className="inline-flex items-center gap-1 rounded-full bg-muted px-3 py-1 text-xs">{trainer.name}<button type="button" aria-label={t('learning_remove_instructor', { name: trainer.name })} className="ml-1 text-destructive" onClick={() => unassignTrainer.mutate(trainer.user_id)}>×</button></span>)}</div>
                        <div className="flex gap-2"><select className="h-9 min-w-0 flex-1 rounded-md border bg-background px-2 text-sm" value={trainerToAssign ?? ''} onChange={(event) => setTrainerToAssign(Number(event.target.value) || null)}><option value="">{t('learning_select_instructor')}</option>{courseTrainersQuery.data?.filter((trainer) => !trainer.assigned).map((trainer) => <option key={trainer.user_id} value={trainer.user_id}>{trainer.name} ({trainer.email})</option>)}</select><Button size="sm" disabled={trainerToAssign === null || assignTrainer.isPending} onClick={() => { if (trainerToAssign !== null) assignTrainer.mutate(trainerToAssign) }}>{t('learning_assign_instructor')}</Button></div>
                      </>}
                    </div>}
                    {courseQuery.isLoading ? <p className="mt-4 text-sm text-muted-foreground">{t('common_loading')}</p> : (
                      <div className="mt-4 space-y-3">
                        {selected.modules?.map((module, moduleIndex) => <div key={module.id} className="rounded-lg bg-muted/40 p-3">
                          <div className="flex items-center justify-between gap-2"><p className="font-medium">{module.position}. {module.title}</p>{canManage && selected.status === 'draft' && <div className="flex items-center">
                            <Button variant="ghost" size="icon" className="size-7" aria-label={t('learning_move_up')} disabled={moduleIndex === 0} onClick={() => { const ids = selected.modules!.map((item) => item.id); [ids[moduleIndex - 1], ids[moduleIndex]] = [ids[moduleIndex], ids[moduleIndex - 1]]; reorderModules.mutate(ids) }}><ArrowUp className="size-3.5" /></Button>
                            <Button variant="ghost" size="icon" className="size-7" aria-label={t('learning_move_down')} disabled={moduleIndex === (selected.modules?.length ?? 0) - 1} onClick={() => { const ids = selected.modules!.map((item) => item.id); [ids[moduleIndex + 1], ids[moduleIndex]] = [ids[moduleIndex], ids[moduleIndex + 1]]; reorderModules.mutate(ids) }}><ArrowDown className="size-3.5" /></Button>
                            <Button variant="ghost" size="icon" className="size-7" aria-label={t('edit')} onClick={() => { const title = window.prompt(t('learning_edit_module'), module.title); if (title?.trim()) updateModule.mutate({ moduleID: module.id, title }) }}><Pencil className="size-3.5" /></Button>
                            <Button variant="ghost" size="icon" className="size-7 text-destructive" aria-label={t('delete')} onClick={() => { if (window.confirm(t('learning_module_delete_confirm', { title: module.title }))) deleteModule.mutate(module.id) }}><Trash2 className="size-3.5" /></Button>
                          </div>}</div>
                          {module.lessons?.map((lesson, lessonIndex) => {
                            const lessonProgress = courseProgressQuery.data?.lessons.find((item) => item.lesson_id === lesson.id)
                            const completedAt = lessonProgress?.completed_at
                            const unlocked = lessonProgress?.unlocked ?? false
                            return <div key={lesson.id} className="ml-4 mt-2 flex flex-wrap items-center justify-between gap-2 text-sm text-muted-foreground"><div className="min-w-0 flex-1"><p className="flex items-center gap-1.5">{isStudent && completedAt && <CheckCircle2 className="size-4 shrink-0 text-primary" />}{isStudent && !unlocked && <LockKeyhole className="size-4 shrink-0" />}{lesson.position}. {lesson.title}</p>{isStudent && unlocked && lesson.content.trim() && <p className="mt-1 whitespace-pre-wrap text-foreground/80">{lesson.content}</p>}{isStudent && lesson.kind === 'task' && unlocked && <TaskSubmissionPanel organizationID={activeOrganizationID!} courseID={selectedCourseID!} lessonID={lesson.id} completed={!!completedAt} onComplete={() => completeLesson.mutate(lesson.id)} />}</div>{isStudent && lesson.kind !== 'task' && <Button type="button" size="sm" variant={completedAt ? 'outline' : 'default'} disabled={!unlocked || !!completedAt || completeLesson.isPending || courseProgressQuery.isError} onClick={() => completeLesson.mutate(lesson.id)}>{t(completedAt ? 'learning_lesson_completed' : unlocked ? 'learning_lesson_complete' : 'learning_lesson_locked')}</Button>}{canManage && selected.status === 'draft' && <div className="flex items-center">
                            <Button variant="ghost" size="icon" className="size-7" aria-label={t('learning_move_up')} disabled={lessonIndex === 0} onClick={() => { const ids = module.lessons!.map((item) => item.id); [ids[lessonIndex - 1], ids[lessonIndex]] = [ids[lessonIndex], ids[lessonIndex - 1]]; reorderLessons.mutate({ moduleID: module.id, ids }) }}><ArrowUp className="size-3.5" /></Button>
                            <Button variant="ghost" size="icon" className="size-7" aria-label={t('learning_move_down')} disabled={lessonIndex === (module.lessons?.length ?? 0) - 1} onClick={() => { const ids = module.lessons!.map((item) => item.id); [ids[lessonIndex + 1], ids[lessonIndex]] = [ids[lessonIndex], ids[lessonIndex + 1]]; reorderLessons.mutate({ moduleID: module.id, ids }) }}><ArrowDown className="size-3.5" /></Button>
                            <Button variant="ghost" size="icon" className="size-7" aria-label={t('edit')} onClick={() => { const title = window.prompt(t('learning_edit_lesson'), lesson.title); if (!title?.trim()) return; const kind = window.prompt(t('learning_edit_lesson_kind'), lesson.kind); if (!kind?.trim()) return; const content = window.prompt(t('learning_edit_lesson_content'), lesson.content); if (content !== null) updateLesson.mutate({ moduleID: module.id, lessonID: lesson.id, title, kind, content }) }}><Pencil className="size-3.5" /></Button>
                            <Button variant="ghost" size="icon" className="size-7 text-destructive" aria-label={t('delete')} onClick={() => { if (window.confirm(t('learning_lesson_delete_confirm', { title: lesson.title }))) deleteLesson.mutate({ moduleID: module.id, lessonID: lesson.id }) }}><Trash2 className="size-3.5" /></Button>
                          </div>}</div>
                          })}
                          {canManage && selected.status === 'draft' && <form className="mt-3 flex flex-wrap gap-2" onSubmit={(event) => { event.preventDefault(); addLesson.mutate({ moduleID: module.id, title: lessonTitle, kind: lessonKind }) }}>
                            <select aria-label={t('learning_lesson_kind')} className="h-8 rounded border bg-background px-2 text-xs" value={lessonKind} onChange={(event) => setLessonKind(event.target.value as 'text' | 'task')}><option value="text">{t('learning_lesson_kind_text')}</option><option value="task">{t('learning_lesson_kind_task')}</option></select>
                            <input required className="h-8 min-w-0 flex-1 rounded border bg-background px-2 text-sm" placeholder={t('learning_lesson_placeholder')} value={lessonModuleID === module.id ? lessonTitle : ''} onFocus={() => { setLessonModuleID(module.id); setLessonTitle('') }} onChange={(event) => { setLessonModuleID(module.id); setLessonTitle(event.target.value) }} />
                            <button className="h-8 rounded border px-2 text-xs">{t('learning_add_lesson')}</button>
                          </form>}
                        </div>)}
                        {canManage && selected.status === 'draft' && <form className="flex gap-2" onSubmit={(event) => { event.preventDefault(); addModule.mutate() }}>
                          <input required className="h-9 min-w-0 flex-1 rounded-md border bg-background px-3 text-sm" placeholder={t('learning_module_name')} value={moduleTitle} onChange={(event) => setModuleTitle(event.target.value)} />
                          <button disabled={addModule.isPending} className="h-9 rounded-md border px-3 text-sm">{t('learning_add_module')}</button>
                        </form>}
                      </div>
                    )}
                  </div>}
                  {canManage && selected?.status === 'published' && activeOrganizationID !== null && <TaskReviewQueue organizationID={activeOrganizationID} courseID={selected.id} />}
                </div>
              </section>
            </>
          )}
        </div>
      </Main>
    </AuthenticatedLayout>
  )
}

function StatCard({ icon, label, value, detail, chart }: { icon: ReactNode; label: string; value: number; detail: string; chart: ReactNode }) {
  return <div className="rounded-xl border bg-card p-5 shadow-sm">
    <div className="flex items-start justify-between gap-4">
      <div className="min-w-0"><p className="text-sm text-muted-foreground">{label}</p><p className="mt-3 text-3xl font-semibold tracking-tight">{value}</p><p className="mt-1 text-xs text-muted-foreground">{detail}</p></div>
      <span className="rounded-lg bg-primary/10 p-2.5 text-primary">{icon}</span>
    </div>
    <div className="mt-5 h-10">{chart}</div>
  </div>
}

function CourseStatusChart({ published, total }: { published: number; total: number }) {
  const publishedPercent = total ? (published / total) * 100 : 0
  return <div role="img" aria-label={`${published} published courses out of ${total}`} className="flex h-full items-center">
    <div className="flex h-2 w-full overflow-hidden rounded-full bg-muted"><div className="h-full rounded-full bg-primary transition-all" style={{ width: `${publishedPercent}%` }} /></div>
  </div>
}

function ProgressRing({ value }: { value: number }) {
  const radius = 15
  const circumference = 2 * Math.PI * radius
  const normalizedValue = Math.min(100, Math.max(0, value))
  return <div className="flex h-full items-center justify-end" role="img" aria-label={`${Math.round(normalizedValue)}% of courses published`}>
    <svg viewBox="0 0 40 40" className="h-10 w-10 -rotate-90">
      <circle cx="20" cy="20" r={radius} fill="none" stroke="currentColor" strokeWidth="5" className="text-muted" />
      <circle cx="20" cy="20" r={radius} fill="none" stroke="currentColor" strokeWidth="5" strokeDasharray={circumference} strokeDashoffset={circumference * (1 - normalizedValue / 100)} strokeLinecap="round" className="text-primary transition-all" />
    </svg>
  </div>
}

function ModuleBars({ value }: { value: number }) {
  const bars = [0.35, 0.58, 0.46, 0.78, 0.63, 1]
  const height = value ? Math.max(30, Math.min(100, value * 18)) : 0
  return <div role="img" aria-label={`${value} modules created`} className="flex h-full items-end justify-end gap-1">
    {bars.map((bar, index) => <span key={index} className={`w-2 rounded-t-sm ${value ? 'bg-primary/70' : 'bg-muted'}`} style={{ height: `${height * bar}%` }} />)}
  </div>
}

function TaskSubmissionPanel({ organizationID, courseID, lessonID, completed, onComplete }: { organizationID: number; courseID: number; lessonID: number; completed: boolean; onComplete: () => void }) {
  const { t } = useI18n()
  const client = useQueryClient()
  const [responseText, setResponseText] = useState('')
  const [responseURL, setResponseURL] = useState('')
  const queryKey = ['learning', 'task-submissions', organizationID, courseID, lessonID]
  const history = useQuery({ queryKey, queryFn: () => learningApi.taskSubmissions(organizationID, courseID, lessonID) })
  const submit = useMutation({
    mutationFn: () => learningApi.submitTask(organizationID, courseID, lessonID, { response_text: responseText, response_url: responseURL }),
    onSuccess: async () => {
      setResponseText('')
      setResponseURL('')
      await client.invalidateQueries({ queryKey })
      onComplete()
      toast.success(t('assessment_submission_sent'))
    },
    onError: () => toast.error(t('assessment_submission_error')),
  })

  return <div className="mt-3 space-y-3 rounded-lg border bg-background p-3">
    <p className="text-sm font-medium text-foreground">{t('assessment_submit_title')}</p>
    <form className="space-y-2" onSubmit={(event) => { event.preventDefault(); submit.mutate() }}>
      <textarea maxLength={20000} className="min-h-24 w-full rounded-md border bg-background px-3 py-2 text-sm text-foreground" placeholder={t('assessment_response_text')} value={responseText} onChange={(event) => setResponseText(event.target.value)} />
      <input type="url" maxLength={2048} className="h-9 w-full rounded-md border bg-background px-3 text-sm text-foreground" placeholder={t('assessment_response_url')} value={responseURL} onChange={(event) => setResponseURL(event.target.value)} />
      <div className="flex flex-wrap items-center justify-between gap-2"><p className="text-xs text-muted-foreground">{t('assessment_multiple_attempts')}</p><Button size="sm" disabled={submit.isPending || (!responseText.trim() && !responseURL.trim())}>{t('assessment_submit')}</Button></div>
    </form>
    {history.isLoading ? <p className="text-xs text-muted-foreground">{t('common_loading')}</p> : history.isError ? <p className="text-xs text-destructive">{t('assessment_history_error')}</p> : history.data?.length ? <div className="space-y-2 border-t pt-3">
      <h4 className="text-sm font-medium text-foreground">{t('assessment_attempt_history')}</h4>
      {history.data.map((submission) => <div key={submission.id} className="rounded-md bg-muted/50 p-3 text-xs text-foreground">
        <div className="flex justify-between gap-2"><span className="font-medium">{t('assessment_attempt', { attempt: submission.attempt })}</span><span>{t(submission.status === 'reviewed' ? 'assessment_reviewed' : 'assessment_waiting_review')}</span></div>
        {submission.response_text && <p className="mt-2 whitespace-pre-wrap">{submission.response_text}</p>}
        {submission.response_url && <a className="mt-1 block break-all text-primary underline" href={submission.response_url} target="_blank" rel="noreferrer">{submission.response_url}</a>}
        {submission.status === 'reviewed' && <div className="mt-2 border-t pt-2"><p className="font-medium">{t('assessment_score')}: {submission.score}</p>{submission.feedback && <p className="mt-1 whitespace-pre-wrap">{submission.feedback}</p>}</div>}
      </div>)}
    </div> : <p className="border-t pt-3 text-xs text-muted-foreground">{t('assessment_no_submissions')}</p>}
    {!completed && !!history.data?.length && <Button type="button" size="sm" variant="outline" onClick={onComplete}>{t('learning_lesson_complete')}</Button>}
  </div>
}

function TaskReviewQueue({ organizationID, courseID }: { organizationID: number; courseID: number }) {
  const { t } = useI18n()
  const submissions = useQuery({
    queryKey: ['learning', 'course-submissions', organizationID, courseID],
    queryFn: () => learningApi.courseSubmissions(organizationID, courseID),
  })

  return <section className="rounded-xl border bg-card p-5 shadow-sm">
    <div className="flex items-center justify-between gap-2"><div><h3 className="font-semibold">{t('assessment_review_queue')}</h3><p className="text-sm text-muted-foreground">{t('assessment_review_queue_description')}</p></div><span className="rounded-full bg-muted px-2.5 py-1 text-xs">{submissions.data?.filter((submission) => submission.status === 'submitted').length ?? 0}</span></div>
    {submissions.isLoading ? <p className="mt-4 text-sm text-muted-foreground">{t('common_loading')}</p> : submissions.isError ? <p className="mt-4 text-sm text-destructive">{t('assessment_review_load_error')}</p> : submissions.data?.length ? <div className="mt-4 space-y-3">{submissions.data.map((submission) => <SubmissionReviewCard key={submission.id} organizationID={organizationID} courseID={courseID} submission={submission} />)}</div> : <p className="mt-4 text-sm text-muted-foreground">{t('assessment_no_submissions')}</p>}
  </section>
}

function SubmissionReviewCard({ organizationID, courseID, submission }: { organizationID: number; courseID: number; submission: AssessmentSubmission }) {
  const { t } = useI18n()
  const client = useQueryClient()
  const [score, setScore] = useState('')
  const [feedback, setFeedback] = useState('')
  const review = useMutation({
    mutationFn: () => learningApi.reviewSubmission(organizationID, courseID, submission.id, { score: Number(score), feedback }),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['learning', 'course-submissions', organizationID, courseID] })
      toast.success(t('assessment_review_saved'))
    },
    onError: () => toast.error(t('assessment_review_error')),
  })

  return <article className="rounded-lg border p-4">
    <div className="flex flex-wrap items-start justify-between gap-2"><div><h4 className="font-medium">{submission.student_name} · {submission.lesson_title}</h4><p className="text-xs text-muted-foreground">{submission.student_email} · {t('assessment_attempt', { attempt: submission.attempt })} · {new Date(submission.submitted_at).toLocaleString()}</p></div><span className="rounded-full bg-muted px-2.5 py-1 text-xs">{t(submission.status === 'reviewed' ? 'assessment_reviewed' : 'assessment_waiting_review')}</span></div>
    {submission.response_text && <p className="mt-3 whitespace-pre-wrap text-sm">{submission.response_text}</p>}
    {submission.response_url && <a className="mt-2 block break-all text-sm text-primary underline" href={submission.response_url} target="_blank" rel="noreferrer">{submission.response_url}</a>}
    {submission.status === 'reviewed' ? <div className="mt-3 border-t pt-3 text-sm"><p className="font-medium">{t('assessment_score')}: {submission.score}</p>{submission.feedback && <p className="mt-1 whitespace-pre-wrap">{submission.feedback}</p>}</div> : <form className="mt-3 grid gap-2 border-t pt-3 sm:grid-cols-[120px_1fr_auto]" onSubmit={(event) => { event.preventDefault(); review.mutate() }}>
      <label className="grid gap-1 text-xs">{t('assessment_score')}<input required type="number" min="0" max="100" step="0.01" className="h-9 rounded-md border bg-background px-2 text-sm" value={score} onChange={(event) => setScore(event.target.value)} /></label>
      <label className="grid gap-1 text-xs">{t('assessment_feedback')}<textarea maxLength={5000} className="min-h-9 rounded-md border bg-background px-2 py-2 text-sm" value={feedback} onChange={(event) => setFeedback(event.target.value)} /></label>
      <Button className="self-end" size="sm" disabled={review.isPending || score === ''}>{t('assessment_review')}</Button>
    </form>}
  </article>
}
