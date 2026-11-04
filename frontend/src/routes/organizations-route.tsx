import { useQuery } from '@tanstack/react-query'
import { useLocation, useParams } from 'react-router-dom'
import OrganizationsPage from '@/pages/organizations'
import CreateOrganizationPage from '@/pages/organizations/create'
import EditOrganizationPage from '@/pages/organizations/edit'
import { RoutePending } from '@/components/route-pending'
import NotFoundError from '@/pages/errors/not-found-error'
import { learningApi } from '@/services/learning-api'

export default function OrganizationsRoute() {
  const { id } = useParams()
  const location = useLocation()
  const isCreate = location.pathname.endsWith('/create')
  const isEdit = !!id && location.pathname.endsWith('/edit')
  const query = useQuery({ queryKey: ['learning', 'organizations'], queryFn: learningApi.organizations, enabled: isEdit })
  if (isCreate) return <CreateOrganizationPage />
  if (isEdit) {
    if (query.isLoading) return <RoutePending />
    const organization = query.data?.find((item) => item.id === Number(id))
    if (!organization) return <NotFoundError />
    return <EditOrganizationPage organization={organization} />
  }
  return <OrganizationsPage />
}
