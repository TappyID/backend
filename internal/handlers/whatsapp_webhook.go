package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tappyone/internal/models"
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

// WebhookPayload representa o payload do webhook WAHA
type WebhookPayload struct {
	Event   string      `json:"event"`
	Session string      `json:"session"`
	Data    interface{} `json:"data"`
}

// SessionStatusData representa dados de mudança de status da sessão
type SessionStatusData struct {
	Status string `json:"status"`
}

func (h *WhatsAppWebhookHandler) ProcessWebhook(c *gin.Context) {
	var payload WebhookPayload
	
	// Parse do JSON payload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Erro ao fazer parse do webhook payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload"})
		return
	}

	log.Printf("Webhook recebido - Event: %s, Session: %s", payload.Event, payload.Session)
	
	// Processar evento de mudança de status da sessão
	if payload.Event == "session.status" {
		h.processSessionStatusChange(payload)
	}
	
	// Processar outras mensagens (se necessário no futuro)
	if payload.Event == "message" {
		log.Printf("Mensagem recebida na sessão %s", payload.Session)
		// Futuro: processar mensagens recebidas
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processado com sucesso"})
}

func (h *WhatsAppWebhookHandler) processSessionStatusChange(payload WebhookPayload) {
	// Parse dos dados do status
	statusData, ok := payload.Data.(map[string]interface{})
	if !ok {
		log.Printf("Erro: dados do status inválidos")
		return
	}

	newStatus, ok := statusData["status"].(string)
	if !ok {
		log.Printf("Erro: status não encontrado nos dados")
		return
	}

	log.Printf("Mudança de status detectada - Sessão: %s, Novo Status: %s", payload.Session, newStatus)

	// Extrair user_id do nome da sessão (formato: user_{uuid})
	if !strings.HasPrefix(payload.Session, "user_") {
		log.Printf("Erro: formato de sessão inválido: %s", payload.Session)
		return
	}
	
	userID := strings.TrimPrefix(payload.Session, "user_")
	log.Printf("Extraído user_id: %s", userID)

	// Atualizar status da conexão WhatsApp no banco
	var connection models.UserConnection
	result := h.db.Where("user_id = ? AND platform = ?", userID, "whatsapp").First(&connection)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			log.Printf("Conexão WhatsApp não encontrada para user_id: %s", userID)
			return
		}
		log.Printf("Erro ao buscar conexão: %v", result.Error)
		return
	}

	// Mapear status WAHA para nosso formato
	var ourStatus models.ConnectionStatus
	switch newStatus {
	case "WORKING":
		ourStatus = models.ConnectionStatusConnected
	case "SCAN_QR_CODE":
		ourStatus = models.ConnectionStatusConnecting
	case "FAILED":
		ourStatus = models.ConnectionStatusError
	case "STOPPED":
		ourStatus = models.ConnectionStatusDisconnected
	default:
		ourStatus = models.ConnectionStatusConnecting
	}

	// Atualizar status se mudou
	if connection.Status != ourStatus {
		log.Printf("Atualizando status da conexão de '%s' para '%s'", connection.Status, ourStatus)
		
		connection.Status = ourStatus
		if err := h.db.Save(&connection).Error; err != nil {
			log.Printf("Erro ao atualizar status da conexão: %v", err)
			return
		}
		
		log.Printf("Status da conexão WhatsApp atualizado com sucesso para: %s", ourStatus)
	} else {
		log.Printf("Status já está atualizado: %s", ourStatus)
	}
}
