package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appclients "boilerplate-learner/internal/application/api_clients"
	"boilerplate-learner/internal/infrastructure/http/httpx"
)

type ExtAuthHandler struct {
	svc *appclients.Service
}

func NewExtAuthHandler(svc *appclients.Service) *ExtAuthHandler {
	return &ExtAuthHandler{svc: svc}
}

type issueTokenRequest struct {
	ClientID     string `json:"client_id"     binding:"required"`
	ClientSecret string `json:"client_secret" binding:"required"`
}

func (h *ExtAuthHandler) Token(c *gin.Context) {
	var req issueTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.IssueToken(c.Request.Context(), req.ClientID, req.ClientSecret)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

type extRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *ExtAuthHandler) Refresh(c *gin.Context) {
	var req extRefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
