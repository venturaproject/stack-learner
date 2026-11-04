import { useState, type FormEvent } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { AtSign, ChevronLeft, GraduationCap, Info, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { AuthenticatedLayout } from '@/layouts'
import { Main } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { useI18n } from '@/i18n/context'
import { learningApi } from '@/services/learning-api'

export default function CreateStudentPage() {
  const { t } = useI18n()
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const client = useQueryClient()
  const organizations = useQuery({ queryKey: ['learning', 'organizations'], queryFn: learningApi.organizations })
  const initialOrg = Number(params.get('organization_id'))
  const [organizationID, setOrganizationID] = useState<number>(initialOrg || 0)
  const [email, setEmail] = useState('')
  const activeOrganizationID = organizationID || organizations.data?.[0]?.id
  const mutation = useMutation({
    mutationFn: () => learningApi.createStudent(activeOrganizationID!, email.trim()),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['learning', 'students', activeOrganizationID] })
      toast.success(t('student_added'))
      navigate(`/admin/students?organization_id=${activeOrganizationID}`)
    },
    onError: () => toast.error(t('student_add_error')),
  })
  const submit = (event: FormEvent<HTMLFormElement>) => { event.preventDefault(); mutation.mutate() }

  const actionBar = <div className="flex justify-end gap-2"><Button type="button" variant="outline" onClick={() => navigate('/admin/students')} disabled={mutation.isPending}>{t('cancel')}</Button><Button type="submit" form="student-enrollment-form" disabled={mutation.isPending || !activeOrganizationID || !email.trim()}>{mutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}{t('student_add')}</Button></div>
  return <AuthenticatedLayout title={t('student_create_title')}><Main><div className="mx-auto grid w-full max-w-6xl flex-1 items-start gap-6">
    <header className="flex items-start gap-4"><Button type="button" variant="outline" size="icon" className="mt-1 shrink-0" onClick={() => navigate('/admin/students')} aria-label={t('back')}><ChevronLeft className="h-4 w-4" /></Button><div><p className="text-sm font-medium text-primary">{t('learning_students_navigation')}</p><h1 className="mt-1 text-2xl font-semibold tracking-tight">{t('student_create_title')}</h1><p className="mt-1 text-sm text-muted-foreground">{t('student_create_description')}</p></div></header>
    {organizations.isLoading ? <p className="text-sm text-muted-foreground">{t('common_loading')}</p> : organizations.isError ? <p className="text-sm text-destructive">{t('organizations_load_error')}</p> : !organizations.data?.length ? <Card className="p-6 text-sm text-muted-foreground">{t('students_need_organization')}</Card> : (
      <div className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
        <form id="student-enrollment-form" onSubmit={submit}>
          <Card className="overflow-hidden"><CardHeader className="border-b bg-muted/20 px-6 py-5"><CardTitle className="text-base">{t('student_enrollment_details')}</CardTitle><CardDescription>{t('student_enrollment_form_description')}</CardDescription></CardHeader><CardContent className="grid gap-6 p-6">
            <div className="grid gap-2"><Label htmlFor="student-organization">{t('learning_organization_label')} <span className="text-destructive">*</span></Label><Select value={activeOrganizationID ? String(activeOrganizationID) : ''} onValueChange={(value) => setOrganizationID(Number(value))}><SelectTrigger id="student-organization" className="h-11"><SelectValue placeholder={t('organization_select_placeholder')} /></SelectTrigger><SelectContent>{organizations.data.map((organization) => <SelectItem key={organization.id} value={String(organization.id)}>{organization.name}</SelectItem>)}</SelectContent></Select><p className="text-xs text-muted-foreground">{t('student_organization_help')}</p></div>
            <div className="grid gap-2"><Label htmlFor="student-email">{t('student_email')} <span className="text-destructive">*</span></Label><div className="relative"><AtSign className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" /><Input id="student-email" type="email" required autoFocus maxLength={254} className="h-11 pl-9" placeholder={t('student_email_placeholder')} value={email} onChange={(event) => setEmail(event.target.value)} /></div><p className="text-xs text-muted-foreground">{t('student_email_help')}</p></div>
          </CardContent></Card>
        </form>
        <aside className="grid gap-4">
          <Card><CardHeader className="pb-3"><div className="flex items-center gap-3"><div className="flex size-10 items-center justify-center rounded-lg bg-primary/10 text-primary"><GraduationCap className="h-5 w-5" /></div><div><CardTitle className="text-base">{t('student_enrollment_preview')}</CardTitle><CardDescription>{t('student_enrollment_preview_caption')}</CardDescription></div></div></CardHeader><CardContent><div className="rounded-lg border bg-muted/20 p-4"><p className="text-sm font-medium">{email || t('student_email_preview')}</p><p className="mt-1 text-xs text-muted-foreground">{organizations.data.find((organization) => organization.id === activeOrganizationID)?.name ?? t('organization_select_placeholder')}</p></div></CardContent></Card>
          <Alert variant="info"><Info className="h-4 w-4" /><AlertTitle>{t('student_existing_account_title')}</AlertTitle><AlertDescription>{t('student_existing_account_help')}</AlertDescription></Alert>
          <div className="hidden lg:block">{actionBar}</div>
        </aside>
        <div className="lg:hidden">{actionBar}</div>
      </div>
    )}
  </div></Main></AuthenticatedLayout>
}
