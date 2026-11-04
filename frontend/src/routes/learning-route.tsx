import { useQuery } from '@tanstack/react-query'
import { useLocation, useParams, useSearchParams } from 'react-router-dom'
import LearningPage from '@/pages/learning'
import CreateCoursePage from '@/pages/learning/course-create'
import EditCoursePage from '@/pages/learning/course-edit'
import { RoutePending } from '@/components/route-pending'
import NotFoundError from '@/pages/errors/not-found-error'
import { learningApi } from '@/services/learning-api'

export default function LearningRoute() {
  const { id } = useParams()
  const location = useLocation()
  const [params] = useSearchParams()
  const isCreate = location.pathname.endsWith('/create')
  const isEdit = !!id && location.pathname.endsWith('/edit')
  const organizationID = Number(params.get('organization_id'))
  const course = useQuery({ queryKey: ['learning', 'course', organizationID, Number(id)], queryFn: () => learningApi.course(organizationID, Number(id)), enabled: isEdit && organizationID > 0 })
  if (isCreate) return <CreateCoursePage />
  if (isEdit) {
    if (!organizationID) return <NotFoundError />
    if (course.isLoading) return <RoutePending />
    if (course.isError || !course.data) return <NotFoundError />
    return <EditCoursePage course={course.data} />
  }
  return <LearningPage />
}
