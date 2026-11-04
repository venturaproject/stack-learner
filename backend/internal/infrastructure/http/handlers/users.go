package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appauth "boilerplate-learner/internal/application/auth"
	"boilerplate-learner/internal/infrastructure/http/httpx"
	"boilerplate-learner/internal/infrastructure/http/pagination"
)

type UserHandler struct {
	svc *appauth.UserService
}

func NewUserHandler(svc *appauth.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// List godoc
// @Summary List users
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page"
// @Param per_page query int false "Per page"
// @Param search query string false "Search"
// @Param status query string false "Status"
// @Success 200 {object} map[string]interface{}
// @Router /users [get]
func (h *UserHandler) List(c *gin.Context) {
	p := pagination.FromContext(c)
	result, err := h.svc.ListUsers(c.Request.Context(), appauth.UserListFilters{
		Search:  queryStr(c, "search"),
		Status:  queryStr(c, "status"),
		Page:    p.Page,
		PerPage: p.PerPage,
	})
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Get godoc
// @Summary Get user by ID
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	user, err := h.svc.GetUser(c.Request.Context(), id)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, formatUser(user))
}

// Create godoc
// @Summary Create user
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body appauth.CreateUserRequest true "User data"
// @Success 201 {object} map[string]interface{}
// @Router /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req appauth.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.svc.CreateUser(c.Request.Context(), req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, formatUser(user))
}

// Update godoc
// @Summary Update user
// @Tags users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body appauth.UpdateUserRequest true "User data"
// @Success 200 {object} map[string]interface{}
// @Router /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var req appauth.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.svc.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, formatUser(user))
}

// Delete godoc
// @Summary Delete user
// @Tags users
// @Security BearerAuth
// @Produce json
// @Param id path int true "User ID"
// @Success 204
// @Router /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := h.svc.DeleteUser(c.Request.Context(), id); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
