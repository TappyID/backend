package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tappyone/internal/services"
)

// WhatsAppWebhookHandler gerencia webhooks do WhatsApp
type WhatsAppWebhookHandler struct {
	db              *gorm.DB
	whatsappService *services.WhatsAppService
}

func NewWhatsAppWebhookHandler(db *gorm.DB, whatsappService *services.WhatsAppService) *WhatsAppWebhookHandler {
	return &WhatsAppWebhookHandler{
		db:              db,
		whatsappService: whatsappService,
	}
}

func (h *WhatsAppWebhookHandler) ProcessWebhook(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Webhook processado"})
}
