package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CobrancaHandler gerencia cobranças
type CobrancaHandler struct {
	db *gorm.DB
}

func NewCobrancaHandler(db *gorm.DB) *CobrancaHandler {
	return &CobrancaHandler{db: db}
}

func (h *CobrancaHandler) GetCobrancas(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Lista de cobranças"})
}

func (h *CobrancaHandler) CreateCobranca(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Cobrança criada"})
}

func (h *CobrancaHandler) GetCobranca(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Buscar cobrança"})
}

func (h *CobrancaHandler) UpdateCobranca(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Cobrança atualizada"})
}

func (h *CobrancaHandler) DeleteCobranca(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Cobrança excluída"})
}
