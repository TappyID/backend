package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ConteudoHandler gerencia conteúdos
type ConteudoHandler struct {
	db *gorm.DB
}

func NewConteudoHandler(db *gorm.DB) *ConteudoHandler {
	return &ConteudoHandler{db: db}
}

func (h *ConteudoHandler) GetConteudos(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Lista de conteúdos"})
}

func (h *ConteudoHandler) CreateConteudo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Conteúdo criado"})
}

func (h *ConteudoHandler) GetConteudo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Buscar conteúdo"})
}

func (h *ConteudoHandler) UpdateConteudo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Conteúdo atualizado"})
}

func (h *ConteudoHandler) DeleteConteudo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Conteúdo excluído"})
}
