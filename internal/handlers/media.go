package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// MediaHandler gerencia mídia
type MediaHandler struct{}

func NewMediaHandler() *MediaHandler {
	return &MediaHandler{}
}

func (h *MediaHandler) UploadMedia(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Upload de mídia"})
}

func (h *MediaHandler) GetMedia(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Buscar mídia"})
}

func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Excluir mídia"})
}
