package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appauth "boilerplate-learner/internal/application/auth"
	"boilerplate-learner/internal/infrastructure/http/httpx"
)

type PermissionHandler struct {
	svc *appauth.PermissionService
}

func NewPermissionHandler(svc *appauth.PermissionService) *PermissionHandler {
	return &PermissionHandler{svc: svc}
}

// List godoc
// @Summary List permissions
// @Tags permissions
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /permissions [get]
func (h *PermissionHandler) List(c *gin.Context) {
	perms, err := h.svc.ListPermissions(c.Request.Context())
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, perms)
}

// Create godoc
// @Summary Create permission
// @Tags permissions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body appauth.CreatePermissionRequest true "Permission data"
// @Success 201 {object} map[string]interface{}
// @Router /permissions [post]
func (h *PermissionHandler) Create(c *gin.Context) {
	var req appauth.CreatePermissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	perm, err := h.svc.CreatePermission(c.Request.Context(), req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, perm)
}

// Delete godoc
// @Summary Delete permission
// @Tags permissions
// @Security BearerAuth
// @Produce json
// @Param id path int true "Permission ID"
// @Success 204
// @Router /permissions/{id} [delete]
func (h *PermissionHandler) Delete(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := h.svc.DeletePermission(c.Request.Context(), id); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
