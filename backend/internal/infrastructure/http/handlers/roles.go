package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appauth "boilerplate-learner/internal/application/auth"
	"boilerplate-learner/internal/infrastructure/http/httpx"
)

type RoleHandler struct {
	svc *appauth.RoleService
}

func NewRoleHandler(svc *appauth.RoleService) *RoleHandler {
	return &RoleHandler{svc: svc}
}

// List godoc
// @Summary List roles
// @Tags roles
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /roles [get]
func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, roles)
}

// Get godoc
// @Summary Get role by ID
// @Tags roles
// @Security BearerAuth
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} map[string]interface{}
// @Router /roles/{id} [get]
func (h *RoleHandler) Get(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	role, err := h.svc.GetRole(c.Request.Context(), id)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, role)
}

// Create godoc
// @Summary Create role
// @Tags roles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body appauth.RoleRequest true "Role data"
// @Success 201 {object} map[string]interface{}
// @Router /roles [post]
func (h *RoleHandler) Create(c *gin.Context) {
	var req appauth.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role, err := h.svc.CreateRole(c.Request.Context(), req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, role)
}

// Update godoc
// @Summary Update role
// @Tags roles
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param body body appauth.RoleRequest true "Role data"
// @Success 200 {object} map[string]interface{}
// @Router /roles/{id} [put]
func (h *RoleHandler) Update(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var req appauth.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	role, err := h.svc.UpdateRole(c.Request.Context(), id, req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, role)
}

// Delete godoc
// @Summary Delete role
// @Tags roles
// @Security BearerAuth
// @Produce json
// @Param id path int true "Role ID"
// @Success 204
// @Router /roles/{id} [delete]
func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := h.svc.DeleteRole(c.Request.Context(), id); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
