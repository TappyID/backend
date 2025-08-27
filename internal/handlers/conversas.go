package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tappyone/internal/services"
)

// ConversaHandler gerencia conversas
type ConversaHandler struct {
	db              *gorm.DB
	whatsappService *services.WhatsAppService
}

func NewConversaHandler(db *gorm.DB, whatsappService *services.WhatsAppService) *ConversaHandler {
	return &ConversaHandler{
		db:              db,
		whatsappService: whatsappService,
	}
}

func (h *ConversaHandler) GetConversas(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Lista de conversas"})
}

func (h *ConversaHandler) CreateConversa(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Conversa criada"})
}

func (h *ConversaHandler) GetConversa(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Buscar conversa"})
}

func (h *ConversaHandler) UpdateConversa(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Conversa atualizada"})
}

func (h *ConversaHandler) DeleteConversa(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Conversa excluída"})
}
