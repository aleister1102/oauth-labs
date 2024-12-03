package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/lab00/client/internal/database"
	"github.com/cyllective/oauth-labs/lab00/client/internal/redis"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// GET /health
func (h *HealthController) Health(c *gin.Context) {
	ctx := c.Request.Context()
	db := database.Get()
	status := "healthy"
	if err := db.PingContext(ctx); err != nil {
		status = "unhealthy"
	}
	rdb := redis.Get()
	if err := rdb.Ping(ctx).Err(); err != nil {
		status = "unhealthy"
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
	})
}
