import { axios } from '@/lib/axios'

export interface AssessmentSubmission {
  id: number
  organization_id: number
  course_id: number
  lesson_id: number
  lesson_title: string
  student_id: number
  student_name: string
  student_email: string
  attempt: number
  response_text: string
  response_url?: string | null
  status: 'submitted' | 'reviewed'
  score?: number | null
  feedback?: string | null
  submitted_at: string
  reviewed_at?: string | null
}

export interface Organization {
  id: number
  name: string
  slug: string
  role: 'admin' | 'trainer' | 'student'
  archived_at?: string | null
  course_count: number
  student_count: number
}

export interface Student {
  user_id: number
  name: string
  email: string
  role: string
}

export interface CourseTrainer {
  user_id: number
  name: string
  email: string
  assigned: boolean
}

export interface Lesson {
  id: number
  module_id: number
  title: string
  kind: string
  content: string
  position: number
}

export interface CourseModule {
  id: number
  course_id: number
  title: string
  position: number
  lessons?: Lesson[]
}

export interface Course {
  id: number
  organization_id: number
  title: string
  slug: string
  description: string
  status: 'draft' | 'published'
  modules?: CourseModule[]
}

export interface Cohort {
  id: number
  organization_id: number
  course_id: number
  course_title: string
  name: string
  starts_at?: string | null
  ends_at?: string | null
  status: 'draft' | 'active' | 'completed' | 'cancelled'
  enrollment_count: number
}

export interface CohortInstructor {
  user_id: number
  name: string
  email: string
  assigned: boolean
}

export interface CohortEnrollment {
  id: number
  cohort_id: number
  user_id: number
  name: string
  email: string
  status: 'active' | 'withdrawn'
  enrolled_at: string
  withdrawn_at?: string | null
  withdrawal_reason?: string | null
}

export interface LessonProgress {
  lesson_id: number
  completed_at?: string | null
  unlocked: boolean
}

export interface CourseProgress {
  course_id: number
  total_lessons: number
  completed_lessons: number
  percent: number
  lessons: LessonProgress[]
}

export interface PageInfo {
  page: number
  per_page: number
  total: number
  total_pages: number
  summary?: { course_count: number; published_count: number; module_count: number }
}

export interface PageResult<T> {
  items: T[]
  pagination: PageInfo
}

export const learningApi = {
  async organizations() {
    const { data } = await axios.get<{ organizations: Organization[] }>('/api/organizations', { params: { per_page: 100 } })
    return data.organizations
  },
  async organizationsPage(page: number, perPage: number, search: string) {
    const { data } = await axios.get<{ organizations: Organization[]; pagination: PageInfo }>('/api/organizations', { params: { page, per_page: perPage, search } })
    return { items: data.organizations, pagination: data.pagination }
  },
  async createOrganization(input: Pick<Organization, 'name' | 'slug'>) {
    const { data } = await axios.post<Organization>('/api/organizations', input)
    return data
  },
  async updateOrganization(id: number, input: Pick<Organization, 'name' | 'slug'>) {
    const { data } = await axios.put<Organization>(`/api/organizations/${id}`, input)
    return data
  },
  async deleteOrganization(id: number) {
    await axios.delete(`/api/organizations/${id}`)
  },
  async restoreOrganization(id: number) {
    await axios.post(`/api/organizations/${id}/restore`)
  },
  async students(organizationID: number) {
    const { data } = await axios.get<{ students: Student[] }>(`/api/organizations/${organizationID}/students`, { params: { per_page: 100 } })
    return data.students
  },
  async studentsPage(organizationID: number, page: number, perPage: number, search: string) {
    const { data } = await axios.get<{ students: Student[]; pagination: PageInfo }>(`/api/organizations/${organizationID}/students`, { params: { page, per_page: perPage, search } })
    return { items: data.students, pagination: data.pagination }
  },
  async createStudent(organizationID: number, email: string) {
    const { data } = await axios.post(`/api/organizations/${organizationID}/students`, { email })
    return data
  },
  async deleteStudent(organizationID: number, studentID: number) {
    await axios.delete(`/api/organizations/${organizationID}/students/${studentID}`)
  },
  async courses(organizationID: number) {
    const { data } = await axios.get<{ courses: Course[] }>(`/api/organizations/${organizationID}/courses`, { params: { per_page: 100 } })
    return data.courses
  },
  async coursesPage(organizationID: number, page: number, perPage: number, search: string) {
    const { data } = await axios.get<{ courses: Course[]; pagination: PageInfo }>(`/api/organizations/${organizationID}/courses`, { params: { page, per_page: perPage, search } })
    return { items: data.courses, pagination: data.pagination }
  },
  async course(organizationID: number, courseID: number) {
    const { data } = await axios.get<Course>(`/api/organizations/${organizationID}/courses/${courseID}`)
    return data
  },
  async courseProgress(organizationID: number, courseID: number) {
    const { data } = await axios.get<CourseProgress>(`/api/organizations/${organizationID}/courses/${courseID}/progress`)
    return data
  },
  async completeLesson(organizationID: number, courseID: number, lessonID: number) {
    const { data } = await axios.post<LessonProgress>(`/api/organizations/${organizationID}/courses/${courseID}/lessons/${lessonID}/complete`)
    return data
  },
  async submitTask(organizationID: number, courseID: number, lessonID: number, input: { response_text: string; response_url: string }) {
    const { data } = await axios.post<AssessmentSubmission>(`/api/organizations/${organizationID}/courses/${courseID}/lessons/${lessonID}/submissions`, input)
    return data
  },
  async taskSubmissions(organizationID: number, courseID: number, lessonID: number) {
    const { data } = await axios.get<{ submissions: AssessmentSubmission[] }>(`/api/organizations/${organizationID}/courses/${courseID}/lessons/${lessonID}/submissions`)
    return data.submissions
  },
  async courseSubmissions(organizationID: number, courseID: number, search = '') {
    const { data } = await axios.get<{ submissions: AssessmentSubmission[]; pagination: PageInfo }>(`/api/organizations/${organizationID}/courses/${courseID}/submissions`, { params: { page: 1, per_page: 100, search } })
    return data.submissions
  },
  async reviewSubmission(organizationID: number, courseID: number, submissionID: number, input: { score: number; feedback: string }) {
    const { data } = await axios.post<AssessmentSubmission>(`/api/organizations/${organizationID}/courses/${courseID}/submissions/${submissionID}/review`, input)
    return data
  },
  async createCourse(organizationID: number, input: Pick<Course, 'title' | 'description'>) {
    const { data } = await axios.post<Course>(`/api/organizations/${organizationID}/courses`, input)
    return data
  },
  async updateCourse(organizationID: number, courseID: number, input: Pick<Course, 'title' | 'description'>) {
    const { data } = await axios.put<Course>(`/api/organizations/${organizationID}/courses/${courseID}`, input)
    return data
  },
  async deleteCourse(organizationID: number, courseID: number) {
    await axios.delete(`/api/organizations/${organizationID}/courses/${courseID}`)
  },
  async publishCourse(organizationID: number, courseID: number) {
    const { data } = await axios.post<Course>(`/api/organizations/${organizationID}/courses/${courseID}/publish`)
    return data
  },
  async courseTrainers(organizationID: number, courseID: number) {
    const { data } = await axios.get<{ trainers: CourseTrainer[] }>(`/api/organizations/${organizationID}/courses/${courseID}/instructors`)
    return data.trainers
  },
  async assignCourseTrainer(organizationID: number, courseID: number, userID: number) {
    await axios.post(`/api/organizations/${organizationID}/courses/${courseID}/instructors`, { user_id: userID })
  },
  async removeCourseTrainer(organizationID: number, courseID: number, userID: number) {
    await axios.delete(`/api/organizations/${organizationID}/courses/${courseID}/instructors/${userID}`)
  },
  async createModule(organizationID: number, courseID: number, title: string) {
    const { data } = await axios.post<CourseModule>(`/api/organizations/${organizationID}/courses/${courseID}/modules`, { title })
    return data
  },
  async updateModule(organizationID: number, courseID: number, moduleID: number, title: string) {
    const { data } = await axios.put<CourseModule>(`/api/organizations/${organizationID}/courses/${courseID}/modules/${moduleID}`, { title })
    return data
  },
  async deleteModule(organizationID: number, courseID: number, moduleID: number) {
    await axios.delete(`/api/organizations/${organizationID}/courses/${courseID}/modules/${moduleID}`)
  },
  async reorderModules(organizationID: number, courseID: number, ids: number[]) {
    await axios.put(`/api/organizations/${organizationID}/courses/${courseID}/modules/order`, { ids })
  },
  async createLesson(organizationID: number, courseID: number, moduleID: number, title: string, kind: string = 'text') {
    const { data } = await axios.post<Lesson>(`/api/organizations/${organizationID}/courses/${courseID}/modules/${moduleID}/lessons`, { title, kind, content: '' })
    return data
  },
  async updateLesson(organizationID: number, courseID: number, moduleID: number, lessonID: number, title: string, kind: string, content: string) {
    const { data } = await axios.put<Lesson>(`/api/organizations/${organizationID}/courses/${courseID}/modules/${moduleID}/lessons/${lessonID}`, { title, kind, content })
    return data
  },
  async deleteLesson(organizationID: number, courseID: number, moduleID: number, lessonID: number) {
    await axios.delete(`/api/organizations/${organizationID}/courses/${courseID}/modules/${moduleID}/lessons/${lessonID}`)
  },
  async reorderLessons(organizationID: number, courseID: number, moduleID: number, ids: number[]) {
    await axios.put(`/api/organizations/${organizationID}/courses/${courseID}/modules/${moduleID}/lessons/order`, { ids })
  },
  async cohorts(organizationID: number, courseID?: number) {
    const path = courseID ? `/api/organizations/${organizationID}/courses/${courseID}/cohorts` : `/api/organizations/${organizationID}/cohorts`
    const { data } = await axios.get<{ cohorts: Cohort[]; pagination: PageInfo }>(path, { params: { page: 1, per_page: 100 } })
    return data.cohorts
  },
  async createCohort(organizationID: number, courseID: number, input: { name: string; starts_at?: string | null; ends_at?: string | null }) {
    const { data } = await axios.post<Cohort>(`/api/organizations/${organizationID}/courses/${courseID}/cohorts`, input)
    return data
  },
  async updateCohort(organizationID: number, cohortID: number, input: { name: string; starts_at?: string | null; ends_at?: string | null }) {
    const { data } = await axios.put<Cohort>(`/api/organizations/${organizationID}/cohorts/${cohortID}`, input)
    return data
  },
  async transitionCohort(organizationID: number, cohortID: number, status: Cohort['status']) {
    await axios.put(`/api/organizations/${organizationID}/cohorts/${cohortID}/status`, { status })
  },
  async cohortInstructors(organizationID: number, cohortID: number) {
    const { data } = await axios.get<{ instructors: CohortInstructor[] }>(`/api/organizations/${organizationID}/cohorts/${cohortID}/instructors`)
    return data.instructors
  },
  async assignCohortInstructor(organizationID: number, cohortID: number, userID: number) {
    await axios.post(`/api/organizations/${organizationID}/cohorts/${cohortID}/instructors`, { user_id: userID })
  },
  async removeCohortInstructor(organizationID: number, cohortID: number, userID: number) {
    await axios.delete(`/api/organizations/${organizationID}/cohorts/${cohortID}/instructors/${userID}`)
  },
  async cohortEnrollments(organizationID: number, cohortID: number, search = '') {
    const { data } = await axios.get<{ enrollments: CohortEnrollment[]; pagination: PageInfo }>(`/api/organizations/${organizationID}/cohorts/${cohortID}/enrollments`, { params: { page: 1, per_page: 100, search } })
    return data.enrollments
  },
  async enrollStudent(organizationID: number, cohortID: number, userID: number) {
    const { data } = await axios.post<CohortEnrollment>(`/api/organizations/${organizationID}/cohorts/${cohortID}/enrollments`, { user_id: userID })
    return data
  },
  async withdrawStudent(organizationID: number, cohortID: number, studentID: number, reason = '') {
    await axios.delete(`/api/organizations/${organizationID}/cohorts/${cohortID}/enrollments/${studentID}`, { data: { reason } })
  },
}
