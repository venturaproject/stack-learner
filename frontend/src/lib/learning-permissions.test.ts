import { describe, expect, it } from 'vitest'
import { canAssignCourseInstructors, canManageCourse, canManageOrganization } from './learning-permissions'

describe('learning permissions for UI affordances', () => {
  it('allows organization management only to organization admins', () => {
    expect(canManageOrganization('admin')).toBe(true)
    expect(canManageOrganization('trainer')).toBe(false)
    expect(canManageOrganization('student')).toBe(false)
    expect(canManageOrganization(undefined)).toBe(false)
  })

  it('allows course management to admins and assigned-course trainers', () => {
    expect(canManageCourse('admin')).toBe(true)
    expect(canManageCourse('trainer')).toBe(true)
    expect(canManageCourse('student')).toBe(false)
  })

  it('restricts instructor assignment to organization admins', () => {
    expect(canAssignCourseInstructors('admin')).toBe(true)
    expect(canAssignCourseInstructors('trainer')).toBe(false)
    expect(canAssignCourseInstructors('student')).toBe(false)
  })
})
