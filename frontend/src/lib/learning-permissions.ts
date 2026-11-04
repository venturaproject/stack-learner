export type OrganizationRole = 'admin' | 'trainer' | 'student' | string

export function canManageOrganization(role: OrganizationRole | undefined): boolean {
  return role === 'admin'
}

export function canManageCourse(role: OrganizationRole | undefined): boolean {
  return role === 'admin' || role === 'trainer'
}

export function canAssignCourseInstructors(role: OrganizationRole | undefined): boolean {
  return role === 'admin'
}
