package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"boilerplate-learner/internal/application/learning"
	learningread "boilerplate-learner/internal/application/learning/read"
	"boilerplate-learner/internal/infrastructure/http/httpx"
	"boilerplate-learner/internal/infrastructure/http/middleware"
	"boilerplate-learner/internal/infrastructure/http/pagination"
)

// LearningHandler exposes organization-scoped LMS operations. Every operation
// receives the authenticated user ID; clients cannot choose a tenant context
// without a matching active membership.
type LearningHandler struct{ service *learning.Service }

func NewLearningHandler(service *learning.Service) *LearningHandler {
	return &LearningHandler{service: service}
}

func (h *LearningHandler) ListOrganizations(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}
	result, err := h.service.ListOrganizations(c.Request.Context(), userID, pageQuery(c))
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"organizations": result.Items, "pagination": result})
}

func (h *LearningHandler) CreateOrganization(c *gin.Context) {
	var in learning.CreateOrganizationInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}
	item, err := h.service.CreateOrganization(c.Request.Context(), userID, in.Name, in.Slug)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, learningread.OrganizationFromDomain(item))
}

func (h *LearningHandler) UpdateOrganization(c *gin.Context) {
	orgID, err := pathInt32(c, "orgID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organización inválida"})
		return
	}
	var in learning.UpdateOrganizationInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}
	item, err := h.service.UpdateOrganization(c.Request.Context(), userID, orgID, in.Name, in.Slug)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, learningread.OrganizationFromDomain(item))
}

func (h *LearningHandler) DeleteOrganization(c *gin.Context) {
	orgID, err := pathInt32(c, "orgID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organización inválida"})
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}
	if err := h.service.ArchiveOrganization(c.Request.Context(), userID, orgID); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LearningHandler) RestoreOrganization(c *gin.Context) {
	orgID, err := pathInt32(c, "orgID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organización inválida"})
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}
	if err := h.service.RestoreOrganization(c.Request.Context(), userID, orgID); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LearningHandler) AddMember(c *gin.Context) {
	orgID, err := pathInt32(c, "orgID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organización inválida"})
		return
	}
	var in learning.AddMemberInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}
	item, err := h.service.AddMember(c.Request.Context(), userID, orgID, in.Email, in.Role)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, learningread.MembershipFromDomain(item))
}

func (h *LearningHandler) ListStudents(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	result, err := h.service.ListStudents(c.Request.Context(), userID, orgID, pageQuery(c))
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"students": result.Items, "pagination": result})
}

func (h *LearningHandler) AddStudent(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	var in learning.AddStudentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.service.AddStudent(c.Request.Context(), userID, orgID, in.Email)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, learningread.MembershipFromDomain(item))
}

func (h *LearningHandler) RemoveStudent(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	studentID, err := pathInt32(c, "studentID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "estudiante inválido"})
		return
	}
	if err := h.service.RemoveStudent(c.Request.Context(), userID, orgID, studentID); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LearningHandler) ListCourses(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	result, err := h.service.ListCourses(c.Request.Context(), userID, orgID, pageQuery(c))
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"courses": learningread.CoursesFromDomain(result.Items), "pagination": result})
}

func pageQuery(c *gin.Context) learning.PageQuery {
	p := pagination.FromContext(c)
	return learning.PageQuery{Page: int(p.Page), PerPage: int(p.PerPage), Search: c.Query("search")}
}

func (h *LearningHandler) CreateCourse(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	var in learning.CreateCourseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.service.CreateCourse(c.Request.Context(), userID, orgID, in.Title, in.Description)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, learningread.CourseFromDomain(item))
}

func (h *LearningHandler) UpdateCourse(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	var in learning.UpdateCourseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.service.UpdateCourse(c.Request.Context(), userID, orgID, courseID, in.Title, in.Description)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, learningread.CourseFromDomain(item))
}

func (h *LearningHandler) DeleteCourse(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	if err := h.service.DeleteCourse(c.Request.Context(), userID, orgID, courseID); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LearningHandler) GetCourse(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	item, err := h.service.GetCourse(c.Request.Context(), userID, orgID, courseID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, learningread.CourseFromDomain(item))
}

func (h *LearningHandler) GetCourseProgress(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	progress, err := h.service.GetCourseProgress(c.Request.Context(), userID, orgID, courseID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, progress)
}

func (h *LearningHandler) CompleteLesson(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	lessonID, err := pathInt32(c, "lessonID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lección inválida"})
		return
	}
	progress, err := h.service.CompleteLesson(c.Request.Context(), userID, orgID, courseID, lessonID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, progress)
}

func (h *LearningHandler) PublishCourse(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	item, err := h.service.PublishCourse(c.Request.Context(), userID, orgID, courseID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, learningread.CourseFromDomain(item))
}

func (h *LearningHandler) AddModule(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	var in learning.TitleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.service.AddModule(c.Request.Context(), userID, orgID, courseID, in.Title)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, learningread.ModuleFromDomain(item))
}

func (h *LearningHandler) AddLesson(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	moduleID, err := pathInt32(c, "moduleID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "módulo inválido"})
		return
	}
	var in learning.AddLessonInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.service.AddLesson(c.Request.Context(), userID, orgID, courseID, moduleID, in.Title, in.Kind, in.Content)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, learningread.LessonFromDomain(item))
}

func (h *LearningHandler) UpdateModule(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	moduleID, err := pathInt32(c, "moduleID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "módulo inválido"})
		return
	}
	var in learning.TitleInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.service.UpdateModule(c.Request.Context(), userID, orgID, courseID, moduleID, in.Title)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, learningread.ModuleFromDomain(item))
}

func (h *LearningHandler) DeleteModule(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	moduleID, err := pathInt32(c, "moduleID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "módulo inválido"})
		return
	}
	if err := h.service.DeleteModule(c.Request.Context(), userID, orgID, courseID, moduleID); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LearningHandler) ReorderModules(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	var in learning.ReorderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	if err := h.service.ReorderModules(c.Request.Context(), userID, orgID, courseID, in.IDs); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LearningHandler) UpdateLesson(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	moduleID, err := pathInt32(c, "moduleID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "módulo inválido"})
		return
	}
	lessonID, err := pathInt32(c, "lessonID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lección inválida"})
		return
	}
	var in learning.AddLessonInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.service.UpdateLesson(c.Request.Context(), userID, orgID, courseID, moduleID, lessonID, in.Title, in.Kind, in.Content)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, learningread.LessonFromDomain(item))
}

func (h *LearningHandler) DeleteLesson(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	moduleID, err := pathInt32(c, "moduleID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "módulo inválido"})
		return
	}
	lessonID, err := pathInt32(c, "lessonID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lección inválida"})
		return
	}
	if err := h.service.DeleteLesson(c.Request.Context(), userID, orgID, courseID, moduleID, lessonID); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LearningHandler) ReorderLessons(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	moduleID, err := pathInt32(c, "moduleID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "módulo inválido"})
		return
	}
	var in learning.ReorderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	if err := h.service.ReorderLessons(c.Request.Context(), userID, orgID, courseID, moduleID, in.IDs); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LearningHandler) ListCourseTrainers(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	items, err := h.service.ListCourseTrainers(c.Request.Context(), userID, orgID, courseID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"trainers": items})
}

func (h *LearningHandler) AssignCourseTrainer(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	var in learning.AssignTrainerInput
	if err := c.ShouldBindJSON(&in); err != nil || in.UserID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "docente inválido"})
		return
	}
	if err := h.service.AssignCourseTrainer(c.Request.Context(), userID, orgID, courseID, in.UserID); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LearningHandler) RemoveCourseTrainer(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	trainerID, err := pathInt32(c, "trainerID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "docente inválido"})
		return
	}
	if err := h.service.RemoveCourseTrainer(c.Request.Context(), userID, orgID, courseID, trainerID); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LearningHandler) ListCohorts(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	var courseID int32
	if c.Param("courseID") != "" {
		parsed, err := pathInt32(c, "courseID")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
			return
		}
		courseID = parsed
	}
	result, err := h.service.ListCohorts(c.Request.Context(), userID, orgID, courseID, pageQuery(c))
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"cohorts": result.Items, "pagination": result})
}

func (h *LearningHandler) CreateCohort(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	courseID, err := pathInt32(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	var in learning.CreateCohortInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.service.CreateCohort(c.Request.Context(), userID, orgID, courseID, in)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *LearningHandler) UpdateCohort(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	cohortID, err := pathInt32(c, "cohortID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cohorte inválida"})
		return
	}
	var in learning.UpdateCohortInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.service.UpdateCohort(c.Request.Context(), userID, orgID, cohortID, in)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *LearningHandler) TransitionCohort(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	cohortID, err := pathInt32(c, "cohortID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cohorte inválida"})
		return
	}
	var in learning.TransitionCohortInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	if err := h.service.TransitionCohort(c.Request.Context(), userID, orgID, cohortID, in.Status); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LearningHandler) ListCohortInstructors(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	cohortID, err := pathInt32(c, "cohortID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cohorte inválida"})
		return
	}
	items, err := h.service.ListCohortInstructors(c.Request.Context(), userID, orgID, cohortID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"instructors": items})
}

func (h *LearningHandler) AssignCohortInstructor(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	cohortID, err := pathInt32(c, "cohortID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cohorte inválida"})
		return
	}
	var in learning.AssignCohortInstructorInput
	if err := c.ShouldBindJSON(&in); err != nil || in.UserID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "docente inválido"})
		return
	}
	if err := h.service.AssignCohortInstructor(c.Request.Context(), userID, orgID, cohortID, in.UserID); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LearningHandler) RemoveCohortInstructor(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	cohortID, err := pathInt32(c, "cohortID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cohorte inválida"})
		return
	}
	trainerID, err := pathInt32(c, "trainerID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "docente inválido"})
		return
	}
	if err := h.service.RemoveCohortInstructor(c.Request.Context(), userID, orgID, cohortID, trainerID); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LearningHandler) ListCohortEnrollments(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	cohortID, err := pathInt32(c, "cohortID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cohorte inválida"})
		return
	}
	result, err := h.service.ListCohortEnrollments(c.Request.Context(), userID, orgID, cohortID, pageQuery(c))
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"enrollments": result.Items, "pagination": result})
}

func (h *LearningHandler) EnrollStudent(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	cohortID, err := pathInt32(c, "cohortID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cohorte inválida"})
		return
	}
	var in learning.EnrollStudentInput
	if err := c.ShouldBindJSON(&in); err != nil || in.UserID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "estudiante inválido"})
		return
	}
	item, err := h.service.EnrollStudent(c.Request.Context(), userID, orgID, cohortID, in.UserID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *LearningHandler) WithdrawStudent(c *gin.Context) {
	orgID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	cohortID, err := pathInt32(c, "cohortID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cohorte inválida"})
		return
	}
	studentID, err := pathInt32(c, "studentID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "estudiante inválido"})
		return
	}
	var in learning.WithdrawStudentInput
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
			return
		}
	}
	if err := h.service.WithdrawStudent(c.Request.Context(), userID, orgID, cohortID, studentID, in.Reason); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *LearningHandler) scope(c *gin.Context) (int32, int32, bool) {
	orgID, err := pathInt32(c, "orgID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organización inválida"})
		return 0, 0, false
	}
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return 0, 0, false
	}
	return orgID, userID, true
}
