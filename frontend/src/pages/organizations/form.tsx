import { useEffect, useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { Building2, Check, ChevronLeft, Globe2, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { AuthenticatedLayout } from '@/layouts'
import { Main } from '@/components/layout'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useI18n } from '@/i18n/context'
import { learningApi, type Organization } from '@/services/learning-api'

export function OrganizationFormPage({ organization }: { organization?: Organization }) {
  const { t } = useI18n()
  const navigate = useNavigate()
  const client = useQueryClient()
  const [name, setName] = useState(organization?.name ?? '')
  const [slug, setSlug] = useState(organization?.slug ?? '')
  useEffect(() => { setName(organization?.name ?? ''); setSlug(organization?.slug ?? '') }, [organization])
  const mutation = useMutation({
    mutationFn: () => organization
      ? learningApi.updateOrganization(organization.id, { name, slug: slug || name })
      : learningApi.createOrganization({ name, slug: slug || name }),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: ['learning', 'organizations'] })
      toast.success(t(organization ? 'organization_updated' : 'organization_created'))
      navigate('/admin/organizations')
    },
    onError: () => toast.error(t('organization_save_error')),
  })
  const submit = (event: FormEvent<HTMLFormElement>) => { event.preventDefault(); mutation.mutate() }
  const editing = !!organization

  const actionLabel = editing ? t('save_changes') : t('organization_add')
  return (
    <AuthenticatedLayout title={t(editing ? 'organization_edit_title' : 'organization_create_title')}>
      <Main><div className="mx-auto grid w-full max-w-6xl flex-1 items-start gap-6">
        <header className="flex items-start gap-4">
          <Button type="button" variant="outline" size="icon" className="mt-1 shrink-0" onClick={() => navigate('/admin/organizations')} aria-label={t('back')}><ChevronLeft className="h-4 w-4" /></Button>
          <div><p className="text-sm font-medium text-primary">{t('learning_organizations_navigation')}</p><h1 className="mt-1 text-2xl font-semibold tracking-tight">{t(editing ? 'organization_edit_title' : 'organization_create_title')}</h1><p className="mt-1 text-sm text-muted-foreground">{t('organization_form_description')}</p></div>
        </header>

        <form onSubmit={submit} className="grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_320px]">
          <Card className="overflow-hidden">
            <CardHeader className="border-b bg-muted/20 px-6 py-5"><CardTitle className="text-base">{t('organization_details')}</CardTitle><CardDescription>{t('organization_form_description')}</CardDescription></CardHeader>
            <CardContent className="grid gap-6 p-6">
              <div className="grid gap-2"><Label htmlFor="organization-name">{t('organization_name')} <span className="text-destructive">*</span></Label><Input id="organization-name" required maxLength={160} autoFocus value={name} onChange={(event) => setName(event.target.value)} placeholder={t('organization_name_placeholder')} className="h-11" /><p className="text-xs text-muted-foreground">{t('organization_name_help')}</p></div>
              <div className="grid gap-2"><Label htmlFor="organization-slug">{t('organization_slug')}</Label><div className="relative"><Globe2 className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" /><Input id="organization-slug" maxLength={100} value={slug} onChange={(event) => setSlug(event.target.value)} placeholder={t('organization_slug_placeholder')} className="h-11 pl-9" /></div><p className="text-xs text-muted-foreground">{t('organization_slug_help')}</p></div>
            </CardContent>
          </Card>

          <aside className="grid gap-4">
            <Card><CardHeader className="pb-3"><div className="flex items-center gap-3"><div className="flex size-10 items-center justify-center rounded-lg bg-primary/10 text-primary"><Building2 className="h-5 w-5" /></div><div className="min-w-0"><CardTitle className="truncate text-base">{name || t('organization_preview')}</CardTitle><CardDescription>{t('organization_preview_caption')}</CardDescription></div></div></CardHeader><CardContent className="space-y-3 text-sm"><div className="flex items-start gap-2 text-muted-foreground"><Check className="mt-0.5 h-4 w-4 shrink-0 text-primary" /><span>{t('organization_preview_courses')}</span></div><div className="flex items-start gap-2 text-muted-foreground"><Check className="mt-0.5 h-4 w-4 shrink-0 text-primary" /><span>{t('organization_preview_students')}</span></div></CardContent></Card>
            <div className="flex justify-end gap-2"><Button type="button" variant="outline" onClick={() => navigate('/admin/organizations')} disabled={mutation.isPending}>{t('cancel')}</Button><Button type="submit" disabled={mutation.isPending || !name.trim()}>{mutation.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}{actionLabel}</Button></div>
          </aside>
        </form>
      </div></Main>
    </AuthenticatedLayout>
  )
}
