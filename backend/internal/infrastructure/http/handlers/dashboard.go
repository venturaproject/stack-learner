package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appdash "boilerplate-learner/internal/application/dashboard"
	"boilerplate-learner/internal/infrastructure/http/httpx"
)

type DashboardHandler struct {
	get *appdash.GetUseCase
}

func NewDashboardHandler(get *appdash.GetUseCase) *DashboardHandler {
	return &DashboardHandler{get: get}
}

// Get godoc
// @Summary Get dashboard stats
// @Tags dashboard
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /dashboard [get]
func (h *DashboardHandler) Get(c *gin.Context) {
	summary, err := h.get.Execute(c.Request.Context())
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	c.JSON(http.StatusOK, summary)
}
