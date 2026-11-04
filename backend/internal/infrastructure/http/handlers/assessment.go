package handlers

import (
	"net/http"
	"strconv"

	appassessment "boilerplate-learner/internal/application/assessment"
	"boilerplate-learner/internal/infrastructure/http/httpx"
	"boilerplate-learner/internal/infrastructure/http/middleware"
	"boilerplate-learner/internal/infrastructure/http/pagination"

	"github.com/gin-gonic/gin"
)

type AssessmentHandler struct{ service *appassessment.Service }

func NewAssessmentHandler(service *appassessment.Service) *AssessmentHandler {
	return &AssessmentHandler{service: service}
}

func (h *AssessmentHandler) Submit(c *gin.Context) {
	orgID, courseID, lessonID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	var input appassessment.SubmitInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.service.Submit(c.Request.Context(), userID, orgID, courseID, lessonID, input)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *AssessmentHandler) ListStudentSubmissions(c *gin.Context) {
	orgID, courseID, lessonID, userID, ok := h.scope(c)
	if !ok {
		return
	}
	items, err := h.service.ListStudentSubmissions(c.Request.Context(), userID, orgID, courseID, lessonID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"submissions": items})
}

func (h *AssessmentHandler) ListCourseSubmissions(c *gin.Context) {
	orgID, userID, ok := h.orgScope(c)
	if !ok {
		return
	}
	courseID, err := parseAssessmentID(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	params := pagination.FromContext(c)
	result, err := h.service.ListCourseSubmissions(c.Request.Context(), userID, orgID, courseID, appassessment.PageQuery{
		Page: int(params.Page), PerPage: int(params.PerPage), Search: c.Query("search"),
	})
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"submissions": result.Items, "pagination": result})
}

func (h *AssessmentHandler) Review(c *gin.Context) {
	orgID, userID, ok := h.orgScope(c)
	if !ok {
		return
	}
	courseID, err := parseAssessmentID(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return
	}
	submissionID, err := strconv.ParseInt(c.Param("submissionID"), 10, 64)
	if err != nil || submissionID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "entrega inválida"})
		return
	}
	var input appassessment.ReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "JSON inválido"})
		return
	}
	item, err := h.service.Review(c.Request.Context(), userID, orgID, courseID, submissionID, input)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *AssessmentHandler) scope(c *gin.Context) (int32, int32, int32, int32, bool) {
	orgID, userID, ok := h.orgScope(c)
	if !ok {
		return 0, 0, 0, 0, false
	}
	courseID, err := parseAssessmentID(c, "courseID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curso inválido"})
		return 0, 0, 0, 0, false
	}
	lessonID, err := parseAssessmentID(c, "lessonID")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lección inválida"})
		return 0, 0, 0, 0, false
	}
	return orgID, courseID, lessonID, userID, true
}

func (h *AssessmentHandler) orgScope(c *gin.Context) (int32, int32, bool) {
	orgID, err := parseAssessmentID(c, "orgID")
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

func parseAssessmentID(c *gin.Context, key string) (int32, error) {
	value, err := strconv.ParseInt(c.Param(key), 10, 32)
	if err != nil || value < 1 {
		return 0, strconv.ErrSyntax
	}
	return int32(value), nil
}
