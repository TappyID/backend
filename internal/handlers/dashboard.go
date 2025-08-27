package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DashboardHandler gerencia dashboard
type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Estatísticas do dashboard"})
}

func (h *DashboardHandler) GetMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Métricas do dashboard"})
}
