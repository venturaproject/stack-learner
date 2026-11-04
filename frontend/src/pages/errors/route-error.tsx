import { isRouteErrorResponse, useRouteError } from 'react-router-dom'
import GeneralError from '@/pages/errors/general-error'
import NotFoundError from '@/pages/errors/not-found-error'
import Forbidden from '@/pages/errors/forbidden'
import UnauthorizedError from '@/pages/errors/unauthorized-error'

export default function RouteError() {
  const error = useRouteError()

  if (isRouteErrorResponse(error)) {
    if (error.status === 404) return <NotFoundError />
    if (error.status === 403) return <Forbidden />
    if (error.status === 401) return <UnauthorizedError />
  }

  return <GeneralError />
}
