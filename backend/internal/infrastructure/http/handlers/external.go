package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appdisp "boilerplate-learner/internal/application/dispositivos"
	dispread "boilerplate-learner/internal/application/dispositivos/read"
	apptelf "boilerplate-learner/internal/application/telefonos"
	telfread "boilerplate-learner/internal/application/telefonos/read"
	apptrab "boilerplate-learner/internal/application/trabajadores"
	trabread "boilerplate-learner/internal/application/trabajadores/read"
	"boilerplate-learner/internal/infrastructure/http/httpx"
	"boilerplate-learner/internal/infrastructure/http/pagination"
)

type ExternalHandler struct {
	trabList *apptrab.ListUseCase
	trabGet  *apptrab.GetUseCase
	telfList *apptelf.ListUseCase
	telfGet  *apptelf.GetUseCase
	dispList *appdisp.ListUseCase
	dispGet  *appdisp.GetUseCase
}

func NewExternalHandler(
	trabList *apptrab.ListUseCase, trabGet *apptrab.GetUseCase,
	telfList *apptelf.ListUseCase, telfGet *apptelf.GetUseCase,
	dispList *appdisp.ListUseCase, dispGet *appdisp.GetUseCase,
) *ExternalHandler {
	return &ExternalHandler{
		trabList: trabList, trabGet: trabGet,
		telfList: telfList, telfGet: telfGet,
		dispList: dispList, dispGet: dispGet,
	}
}

// ── Trabajadores ─────────────────────────────────────────────────────────────

func (h *ExternalHandler) ListTrabajadores(c *gin.Context) {
	p := pagination.FromContext(c)
	result, err := h.trabList.Execute(c.Request.Context(), trabread.Filters{
		Search:    queryStr(c, "search"),
		Activo:    queryStr(c, "activo"),
		Ubicacion: queryStr(c, "ubicacion"),
		Ciudad:    queryStr(c, "ciudad"),
		Page:      p.Page,
		PerPage:   p.PerPage,
	})
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"results":  result.Results,
		"total":    result.Total,
		"page":     result.Page,
		"per_page": result.PerPage,
	})
}

func (h *ExternalHandler) GetTrabajador(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	d, err := h.trabGet.Execute(c.Request.Context(), id)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}

// ── Teléfonos ────────────────────────────────────────────────────────────────

func (h *ExternalHandler) ListTelefonos(c *gin.Context) {
	p := pagination.FromContext(c)
	result, err := h.telfList.Execute(c.Request.Context(), telfread.Filters{
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
	c.JSON(http.StatusOK, gin.H{
		"results":  result.Results,
		"total":    result.Total,
		"page":     result.Page,
		"per_page": result.PerPage,
	})
}

func (h *ExternalHandler) GetTelefono(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	d, err := h.telfGet.Execute(c.Request.Context(), id)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}

// ── Dispositivos ─────────────────────────────────────────────────────────────

func (h *ExternalHandler) ListDispositivos(c *gin.Context) {
	p := pagination.FromContext(c)
	result, err := h.dispList.Execute(c.Request.Context(), dispread.Filters{
		Search:  queryStr(c, "search"),
		Estado:  queryStr(c, "estado"),
		Grupo:   queryStr(c, "grupo"),
		Page:    p.Page,
		PerPage: p.PerPage,
	})
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"results":  result.Results,
		"total":    result.Total,
		"page":     result.Page,
		"per_page": result.PerPage,
	})
}

func (h *ExternalHandler) GetDispositivo(c *gin.Context) {
	id, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	d, err := h.dispGet.Execute(c.Request.Context(), id)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}
