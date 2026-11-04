import { useState } from 'react'
import { AlertCircle } from 'lucide-react'
import { DocumentTitle } from '@/components/document-title'
import { Alert, AlertTitle, AlertDescription } from '@/components/ui/alert'
import { UserAuthForm } from './components/user-auth-form'
import { useI18n } from '@/i18n/context'
import { consumeSessionExpired } from '@/lib/session-expiry'

export default function SignIn2({
    status,
    canResetPassword: _canResetPassword,
  }: {
  status?: string;
  canResetPassword: boolean;
}) {
  const { t } = useI18n()
  // Read once per mount (not per render) — consumeSessionExpired() clears the
  // flag as a side effect, so calling it directly in the render body would
  // make the banner disappear on the very next re-render (e.g. typing in the
  // form) instead of staying up for the whole visit to this page.
  const [sessionExpired] = useState(consumeSessionExpired)
  return (
    <>
      <DocumentTitle title={t('login')}/>
      <div
        className='relative flex min-h-svh flex-col items-center justify-center px-4 lg:max-w-none lg:grid lg:grid-cols-2 lg:px-0'>
        <div className='relative hidden h-full flex-col bg-muted p-10 text-white dark:border-r lg:flex'>
          <div className='absolute inset-0 bg-zinc-900'/>
          <div className='relative z-20 flex items-center text-lg font-medium'>
          </div>

          <img
            src='/learner-logo-compact-dark.svg'
            className='relative m-auto'
            width={301}
            height={60}
            alt='Learner'
          />

          <div className='relative z-20 mt-auto'>
            <blockquote className='space-y-2'>
              <footer className='text-sm text-center'>
                {import.meta.env.VITE_APP_NAME || 'Learner'} © {new Date().getFullYear()}
              </footer>
            </blockquote>
          </div>
        </div>
        <div className='w-full lg:p-8'>
          <div className='mx-auto flex w-full flex-col justify-center space-y-6 sm:w-[350px]'>
            <div className='flex flex-col items-center space-y-2 text-center lg:items-start lg:text-left'>
              <img
                src='/learner-logo-compact-light.svg'
                className='mb-4 h-10 w-10 lg:hidden'
                alt='Learner'
              />
              <h1 className='text-2xl font-semibold tracking-tight'>Iniciar sesión</h1>
              <p className='text-sm text-muted-foreground'>
                {t('enter_email_and_password_below')} <br className="hidden sm:inline"/>
                {t('to_log_into_your_account')}
              </p>
            </div>
            {sessionExpired && (
              <Alert variant="info">
                <AlertCircle className="h-4 w-4" />
                <AlertTitle>{t('auth_session_expired_title')}</AlertTitle>
                <AlertDescription>{t('auth_session_expired_message')}</AlertDescription>
              </Alert>
            )}
            <UserAuthForm canResetPassword={false} status={status} />
          </div>
        </div>
      </div>
    </>
  )
}
