import { CourseFormPage } from './course-form'
import type { Course } from '@/services/learning-api'

export default function EditCoursePage({ course }: { course: Course }) {
  return <CourseFormPage course={course} />
}
