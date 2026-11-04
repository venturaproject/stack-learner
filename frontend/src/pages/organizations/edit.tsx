import { OrganizationFormPage } from './form'
import type { Organization } from '@/services/learning-api'

export default function EditOrganizationPage({ organization }: { organization: Organization }) {
  return <OrganizationFormPage organization={organization} />
}
