package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apptrab "boilerplate-learner/internal/application/trabajadores"
	trabread "boilerplate-learner/internal/application/trabajadores/read"
	"boilerplate-learner/internal/infrastructure/http/httpx"
	"boilerplate-learner/internal/infrastructure/http/pagination"
)

type TrabajadorHandler struct {
	uc *apptrab.UseCases
}

func NewTrabajadorHandler(uc *apptrab.UseCases) *TrabajadorHandler {
	return &TrabajadorHandler{uc: uc}
}

// List godoc
// @Summary List trabajadores
// @Tags trabajadores
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /trabajadores [get]
func (h *TrabajadorHandler) List(c *gin.Context) {
	p := pagination.FromContext(c)
	result, err := h.uc.List.Execute(c.Request.Context(), trabread.Filters{
		Search:    queryStr(c, "search"),
		Activo:    queryStr(c, "activo"),
		Ubicacion: queryStr(c, "ubicacion"),
		Ciudad:    queryStr(c, "ciudad"),
		DateFrom:  queryStr(c, "date_from"),
		DateTo:    queryStr(c, "date_to"),
		Page:      p.Page,
		PerPage:   p.PerPage,
	})
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Get godoc
// @Summary Get trabajador by ID
// @Tags trabajadores
// @Security BearerAuth
// @Produce json
// @Param id path int true "Trabajador ID"
// @Success 200 {object} map[string]interface{}
// @Router /trabajadores/{id} [get]
func (h *TrabajadorHandler) Get(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	d, err := h.uc.Get.Execute(c.Request.Context(), id)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}

// Create godoc
// @Summary Create trabajador
// @Tags trabajadores
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body apptrab.CreateInput true "Trabajador data"
// @Success 201 {object} map[string]interface{}
// @Router /trabajadores [post]
func (h *TrabajadorHandler) Create(c *gin.Context) {
	var in apptrab.CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	created, err := h.uc.Create.Execute(c.Request.Context(), in)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, created)
}

// Update godoc
// @Summary Update trabajador
// @Tags trabajadores
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Trabajador ID"
// @Param body body apptrab.UpdateInput true "Trabajador data"
// @Success 200 {object} map[string]interface{}
// @Router /trabajadores/{id} [put]
func (h *TrabajadorHandler) Update(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var in apptrab.UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.uc.Update.Execute(c.Request.Context(), id, in)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

// Delete godoc
// @Summary Delete trabajador
// @Tags trabajadores
// @Security BearerAuth
// @Param id path int true "Trabajador ID"
// @Success 204
// @Router /trabajadores/{id} [delete]
func (h *TrabajadorHandler) Delete(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := h.uc.Delete.Execute(c.Request.Context(), id); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// DarDeBaja godoc
// @Summary Dar de baja a un trabajador
// @Tags trabajadores
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Trabajador ID"
// @Success 200 {object} map[string]interface{}
// @Router /trabajadores/{id}/dar-de-baja [post]
func (h *TrabajadorHandler) DarDeBaja(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var body struct {
		FechaBaja *string `json:"fecha_baja"`
	}
	_ = c.ShouldBindJSON(&body)
	updated, err := h.uc.DarDeBaja.Execute(c.Request.Context(), id, body.FechaBaja)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}
