package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AutomacaoHandler gerencia automações
type AutomacaoHandler struct {
	db *gorm.DB
}

func NewAutomacaoHandler(db *gorm.DB) *AutomacaoHandler {
	return &AutomacaoHandler{db: db}
}

func (h *AutomacaoHandler) GetAutomacoes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Lista de automações"})
}

func (h *AutomacaoHandler) CreateAutomacao(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Automação criada"})
}

func (h *AutomacaoHandler) GetAutomacao(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Buscar automação"})
}

func (h *AutomacaoHandler) UpdateAutomacao(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Automação atualizada"})
}

func (h *AutomacaoHandler) DeleteAutomacao(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Automação excluída"})
}
