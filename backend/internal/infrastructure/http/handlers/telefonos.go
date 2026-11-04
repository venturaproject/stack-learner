package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apptelf "boilerplate-learner/internal/application/telefonos"
	telfread "boilerplate-learner/internal/application/telefonos/read"
	"boilerplate-learner/internal/infrastructure/http/httpx"
	"boilerplate-learner/internal/infrastructure/http/middleware"
	"boilerplate-learner/internal/infrastructure/http/pagination"
)

type TelefonoHandler struct {
	uc *apptelf.UseCases
}

func NewTelefonoHandler(uc *apptelf.UseCases) *TelefonoHandler {
	return &TelefonoHandler{uc: uc}
}

// List godoc
// @Summary List teléfonos
// @Tags telefonos
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /telefonos [get]
func (h *TelefonoHandler) List(c *gin.Context) {
	p := pagination.FromContext(c)
	result, err := h.uc.List.Execute(c.Request.Context(), telfread.Filters{
		Search:  queryStr(c, "search"),
		Status:  queryStr(c, "status"),
		Plan:    queryStr(c, "plan"),
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
// @Summary Get teléfono by ID
// @Tags telefonos
// @Security BearerAuth
// @Produce json
// @Param id path int true "Teléfono ID"
// @Success 200 {object} map[string]interface{}
// @Router /telefonos/{id} [get]
func (h *TelefonoHandler) Get(c *gin.Context) {
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

// ListEstados godoc
// @Summary List estados de teléfono
// @Tags telefonos
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /telefonos/estados [get]
func (h *TelefonoHandler) ListEstados(c *gin.Context) {
	estados, err := h.uc.ListEstados.Execute(c.Request.Context())
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, estados)
}

// ListTipologias godoc
// @Summary List tipologías
// @Tags telefonos
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /telefonos/tipologias [get]
func (h *TelefonoHandler) ListTipologias(c *gin.Context) {
	tipologias, err := h.uc.ListTipologias.Execute(c.Request.Context())
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, tipologias)
}

// Create godoc
// @Summary Create teléfono
// @Tags telefonos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body apptelf.CreateInput true "Teléfono data"
// @Success 201 {object} map[string]interface{}
// @Router /telefonos [post]
func (h *TelefonoHandler) Create(c *gin.Context) {
	var in apptelf.CreateInput
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
// @Summary Update teléfono
// @Tags telefonos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Teléfono ID"
// @Param body body apptelf.UpdateInput true "Teléfono data"
// @Success 200 {object} map[string]interface{}
// @Router /telefonos/{id} [put]
func (h *TelefonoHandler) Update(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var in apptelf.UpdateInput
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
// @Summary Delete teléfono
// @Tags telefonos
// @Security BearerAuth
// @Param id path int true "Teléfono ID"
// @Success 204
// @Router /telefonos/{id} [delete]
func (h *TelefonoHandler) Delete(c *gin.Context) {
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

// Desactivar godoc
// @Summary Desactivar teléfono
// @Tags telefonos
// @Security BearerAuth
// @Param id path int true "Teléfono ID"
// @Success 200 {object} map[string]interface{}
// @Router /telefonos/{id}/desactivar [post]
func (h *TelefonoHandler) Desactivar(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	updated, err := h.uc.Desactivar.Execute(c.Request.Context(), id)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

// Asignar godoc
// @Summary Asignar teléfono a trabajador
// @Tags telefonos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Teléfono ID"
// @Param body body apptelf.AsignarInput true "Asignación"
// @Success 200 {object} map[string]interface{}
// @Router /telefonos/{id}/asignar [post]
func (h *TelefonoHandler) Asignar(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var in apptelf.AsignarInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.uc.Asignar.Execute(c.Request.Context(), id, in)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

// CreateActuacion godoc
// @Summary Crear actuación en teléfono
// @Tags telefonos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Teléfono ID"
// @Param body body apptelf.ActuacionInput true "Actuación"
// @Success 201 {object} map[string]interface{}
// @Router /telefonos/{id}/actuaciones [post]
func (h *TelefonoHandler) CreateActuacion(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var in apptelf.ActuacionInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := middleware.GetUserID(c)
	var userIDPtr *int32
	if userID != 0 {
		userIDPtr = &userID
	}
	actuacion, err := h.uc.AddActuacion.Execute(c.Request.Context(), id, in, userIDPtr)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, actuacion)
}

// DeleteActuacion godoc
// @Summary Eliminar actuación
// @Tags telefonos
// @Security BearerAuth
// @Param id path int true "Actuación ID"
// @Success 204
// @Router /telefonos/actuaciones/{id} [delete]
func (h *TelefonoHandler) DeleteActuacion(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	if err := h.uc.RemoveActuacion.Execute(c.Request.Context(), id); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
