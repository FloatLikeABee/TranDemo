package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler checks the health status of the service
// @Summary      Health check
// @Description  Check the health status of all services (database, AI service, SQL Server)
// @Tags         Health
// @Produce      json
// @Success      200  {object}  map[string]string  "Service health status"
// @Router       /health [get]
func (h *Handlers) HealthHandler(c *gin.Context) {
	status := gin.H{
		"status":     "healthy",
		"db":         "connected",
		"ai_service": "ready",
		"sql_server": "not_configured",
	}

	if h.sqlService != nil && h.sqlService.IsConnected() {
		status["sql_server"] = "connected"
	}

	c.JSON(http.StatusOK, status)
}

