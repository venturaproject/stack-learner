package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appclients "boilerplate-learner/internal/application/api_clients"
	"boilerplate-learner/internal/infrastructure/http/httpx"
)

type ApiClientHandler struct {
	svc *appclients.Service
}

func NewApiClientHandler(svc *appclients.Service) *ApiClientHandler {
	return &ApiClientHandler{svc: svc}
}

func (h *ApiClientHandler) Svc() *appclients.Service { return h.svc }

func (h *ApiClientHandler) List(c *gin.Context) {
	dtos, err := h.svc.List(c.Request.Context())
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, dtos)
}

type createApiClientRequest struct {
	Name   string   `json:"name"   binding:"required"`
	Scopes []string `json:"scopes" binding:"required,min=1"`
}

func (h *ApiClientHandler) Create(c *gin.Context) {
	var req createApiClientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.Create(c.Request.Context(), appclients.CreateRequest{
		Name:   req.Name,
		Scopes: req.Scopes,
	})
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"client": resp.Client,
		"secret": resp.RawSecret,
	})
}

func (h *ApiClientHandler) Revoke(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Revoke(c.Request.Context(), id); err != nil {
		httpx.Respond(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
