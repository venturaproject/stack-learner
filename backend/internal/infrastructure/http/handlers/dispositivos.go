package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	appdisp "boilerplate-learner/internal/application/dispositivos"
	dispread "boilerplate-learner/internal/application/dispositivos/read"
	"boilerplate-learner/internal/infrastructure/http/httpx"
	"boilerplate-learner/internal/infrastructure/http/pagination"
)

type DispositivoHandler struct {
	uc *appdisp.UseCases
}

func NewDispositivoHandler(uc *appdisp.UseCases) *DispositivoHandler {
	return &DispositivoHandler{uc: uc}
}

// List godoc
// @Summary List dispositivos
// @Tags dispositivos
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /dispositivos [get]
func (h *DispositivoHandler) List(c *gin.Context) {
	p := pagination.FromContext(c)

	var trabID *int32
	if v := c.Query("trabajador_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 32); err == nil {
			id32 := int32(id)
			trabID = &id32
		}
	}

	result, err := h.uc.List.Execute(c.Request.Context(), dispread.Filters{
		Search:       queryStr(c, "search"),
		Marca:        queryStr(c, "marca"),
		Grupo:        queryStr(c, "grupo"),
		Estado:       queryStr(c, "estado"),
		TrabajadorID: trabID,
		Page:         p.Page,
		PerPage:      p.PerPage,
	})
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Get godoc
// @Summary Get dispositivo by ID
// @Tags dispositivos
// @Security BearerAuth
// @Produce json
// @Param id path int true "Dispositivo ID"
// @Success 200 {object} map[string]interface{}
// @Router /dispositivos/{id} [get]
func (h *DispositivoHandler) Get(c *gin.Context) {
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
// @Summary Create dispositivo
// @Tags dispositivos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body appdisp.CreateInput true "Dispositivo data"
// @Success 201 {object} map[string]interface{}
// @Router /dispositivos [post]
func (h *DispositivoHandler) Create(c *gin.Context) {
	var in appdisp.CreateInput
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
// @Summary Update dispositivo
// @Tags dispositivos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Dispositivo ID"
// @Param body body appdisp.UpdateInput true "Dispositivo data"
// @Success 200 {object} map[string]interface{}
// @Router /dispositivos/{id} [put]
func (h *DispositivoHandler) Update(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var in appdisp.UpdateInput
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
// @Summary Delete dispositivo
// @Tags dispositivos
// @Security BearerAuth
// @Param id path int true "Dispositivo ID"
// @Success 204
// @Router /dispositivos/{id} [delete]
func (h *DispositivoHandler) Delete(c *gin.Context) {
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

// Entregar godoc
// @Summary Entregar dispositivo a trabajador
// @Tags dispositivos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Dispositivo ID"
// @Param body body appdisp.EntregarInput true "Entrega"
// @Success 200 {object} map[string]interface{}
// @Router /dispositivos/{id}/entregar [post]
func (h *DispositivoHandler) Entregar(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var in appdisp.EntregarInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updated, err := h.uc.Entregar.Execute(c.Request.Context(), id, in)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, updated)
}

// ── Catalog ──────────────────────────────────────────────────────────────────

func (h *DispositivoHandler) ListBrands(c *gin.Context) {
	brands, err := h.uc.ListBrands.Execute(c.Request.Context())
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, brands)
}

func (h *DispositivoHandler) CreateBrand(c *gin.Context) {
	var in appdisp.BrandInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	brand, err := h.uc.CreateBrand.Execute(c.Request.Context(), in)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, brand)
}

func (h *DispositivoHandler) UpdateBrand(c *gin.Context) {
	id, _ := pathInt32(c, "id")
	var in appdisp.BrandInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	brand, err := h.uc.UpdateBrand.Execute(c.Request.Context(), id, in)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, brand)
}

func (h *DispositivoHandler) DeleteBrand(c *gin.Context) {
	id, _ := pathInt32(c, "id")
	if err := h.uc.DeleteBrand.Execute(c.Request.Context(), id); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *DispositivoHandler) ListModels(c *gin.Context) {
	var brandID *int32
	if v := c.Query("brand_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 32); err == nil {
			id32 := int32(id)
			brandID = &id32
		}
	}
	models, err := h.uc.ListModels.Execute(c.Request.Context(), brandID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, models)
}

func (h *DispositivoHandler) CreateModel(c *gin.Context) {
	var in appdisp.ModelInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	model, err := h.uc.CreateModel.Execute(c.Request.Context(), in)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusCreated, model)
}

func (h *DispositivoHandler) UpdateModel(c *gin.Context) {
	id, _ := pathInt32(c, "id")
	var in appdisp.ModelInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	model, err := h.uc.UpdateModel.Execute(c.Request.Context(), id, in)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, model)
}

func (h *DispositivoHandler) DeleteModel(c *gin.Context) {
	id, _ := pathInt32(c, "id")
	if err := h.uc.DeleteModel.Execute(c.Request.Context(), id); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
