package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// QuadroHandler gerencia quadros
type QuadroHandler struct {
	db *gorm.DB
}

func NewQuadroHandler(db *gorm.DB) *QuadroHandler {
	return &QuadroHandler{db: db}
}

func (h *QuadroHandler) GetQuadros(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Lista de quadros"})
}

func (h *QuadroHandler) CreateQuadro(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Quadro criado"})
}

func (h *QuadroHandler) GetQuadro(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Buscar quadro"})
}

func (h *QuadroHandler) UpdateQuadro(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Quadro atualizado"})
}

func (h *QuadroHandler) DeleteQuadro(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Quadro excluído"})
}
