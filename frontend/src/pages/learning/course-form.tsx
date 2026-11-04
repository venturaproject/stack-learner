import { useEffect, useState, type FormEvent } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ChevronLeft, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { AuthenticatedLayout } from '@/layouts'
import { Main } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { useI18n } from '@/i18n/context'
import { learningApi, type Course } from '@/services/learning-api'

export function CourseFormPage({ course }: { course?: Course }) {
  const { t } = useI18n()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const client = useQueryClient()
  const organizations = useQuery({ queryKey: ['learning', 'organizations'], queryFn: learningApi.organizations })
  const [organizationID, setOrganizationID] = useState<number>(() => course?.organization_id ?? (Number(params.get('organization_id')) || 0))
  const [title, setTitle] = useState(course?.title ?? '')
  const [description, setDescription] = useState(course?.description ?? '')
  useEffect(() => { if (course) { setTitle(course.title); setDescription(course.description); setOrganizationID(course.organization_id) } }, [course])
  const selectedOrganizationID = organizationID || organizations.data?.[0]?.id
  const editing = !!course
  const mutation = useMutation({
    mutationFn: () => editing
      ? learningApi.updateCourse(course.organization_id, course.id, { title, description })
      : learningApi.createCourse(selectedOrganizationID!, { title, description }),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['learning'] })
      toast.success(t(editing ? 'learning_course_updated' : 'learning_course_created_draft'))
      navigate(`/admin/learning?organization_id=${editing ? course.organization_id : selectedOrganizationID}`)
    },
    onError: () => toast.error(t(editing ? 'learning_course_update_error' : 'learning_course_create_error')),
  })
  const submit = (event: FormEvent<HTMLFormElement>) => { event.preventDefault(); mutation.mutate() }

  return <AuthenticatedLayout title={t(editing ? 'learning_edit_course' : 'learning_new_course')}><Main><div className="grid flex-1 items-start gap-4 md:gap-8">
    <form onSubmit={submit} className="grid flex-1 auto-rows-max gap-4">
      <div className="flex items-center gap-4"><Button type="button" variant="outline" size="icon" className="h-8 w-8" onClick={() => navigate('/admin/learning')} aria-label={t('back')}><ChevronLeft className="h-4 w-4" /></Button><h1 className="flex-1 text-xl font-semibold tracking-tight">{t(editing ? 'learning_edit_course' : 'learning_new_course')}</h1><div className="ml-auto hidden gap-2 md:flex"><Button type="button" variant="outline" onClick={() => navigate('/admin/learning')}>{t('cancel')}</Button><Button type="submit" disabled={mutation.isPending || !selectedOrganizationID}>{mutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}{t(editing ? 'save_changes' : 'learning_create_draft')}</Button></div></div>
      <Card><CardHeader><CardTitle>{t('learning_course_details')}</CardTitle><CardDescription>{t('learning_course_form_description')}</CardDescription></CardHeader><CardContent className="grid gap-6">
        {!editing && <div className="grid max-w-xl gap-2"><Label>{t('learning_organization_label')}</Label><Select value={selectedOrganizationID ? String(selectedOrganizationID) : ''} onValueChange={(value) => setOrganizationID(Number(value))}><SelectTrigger><SelectValue placeholder={t('organization_select_placeholder')} /></SelectTrigger><SelectContent>{organizations.data?.map((organization) => <SelectItem key={organization.id} value={String(organization.id)}>{organization.name}</SelectItem>)}</SelectContent></Select></div>}
        <div className="grid gap-2"><Label htmlFor="course-title">{t('learning_course_name')}</Label><Input id="course-title" required maxLength={160} value={title} onChange={(event) => setTitle(event.target.value)} /></div>
        <div className="grid gap-2"><Label htmlFor="course-description">{t('learning_course_description')}</Label><textarea id="course-description" maxLength={5000} rows={5} className="flex min-h-24 w-full rounded-md border border-input bg-background px-3 py-2 text-sm" value={description} onChange={(event) => setDescription(event.target.value)} /></div>
      </CardContent></Card>
      <div className="flex justify-end gap-2 md:hidden"><Button type="button" variant="outline" onClick={() => navigate('/admin/learning')}>{t('cancel')}</Button><Button type="submit" disabled={mutation.isPending || !selectedOrganizationID}>{t(editing ? 'save_changes' : 'learning_create_draft')}</Button></div>
    </form>
  </div></Main></AuthenticatedLayout>
}
