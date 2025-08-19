package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"tappyone/internal/services"
)

// WhatsAppMessageHandler gerencia ações de mensagens do WhatsApp
type WhatsAppMessageHandler struct {
	whatsappService *services.WhatsAppService
}

func NewWhatsAppMessageHandler(whatsappService *services.WhatsAppService) *WhatsAppMessageHandler {
	return &WhatsAppMessageHandler{whatsappService: whatsappService}
}

// ReplyMessage responde a uma mensagem
func (h *WhatsAppMessageHandler) ReplyMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sessionName := fmt.Sprintf("user_%s", userID)
	chatID := c.Param("chatId")

	var req struct {
		MessageID string `json:"messageId" binding:"required"`
		Text      string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[WHATSAPP] ReplyMessage - SessionName: %s, ChatID: %s, MessageID: %s", sessionName, chatID, req.MessageID)

	_, err := h.whatsappService.SendReplyMessage(sessionName, chatID, req.Text, req.MessageID)
	if err != nil {
		log.Printf("[WHATSAPP] ReplyMessage - Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao responder mensagem"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mensagem respondida com sucesso"})
}

// ForwardMessage encaminha uma mensagem
func (h *WhatsAppMessageHandler) ForwardMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sessionName := fmt.Sprintf("user_%s", userID)

	var req struct {
		MessageID string `json:"messageId" binding:"required"`
		ToChatID  string `json:"toChatId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[WHATSAPP] ForwardMessage - SessionName: %s, MessageID: %s, ToChatID: %s", sessionName, req.MessageID, req.ToChatID)

	_, err := h.whatsappService.ForwardMessage(sessionName, req.ToChatID, req.MessageID)
	if err != nil {
		log.Printf("[WHATSAPP] ForwardMessage - Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao encaminhar mensagem"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mensagem encaminhada com sucesso"})
}

// EditMessage edita uma mensagem
func (h *WhatsAppMessageHandler) EditMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sessionName := fmt.Sprintf("user_%s", userID)
	chatID := c.Param("chatId")
	messageID := c.Param("messageId")

	var req struct {
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[WHATSAPP] EditMessage - SessionName: %s, ChatID: %s, MessageID: %s", sessionName, chatID, messageID)

	_, err := h.whatsappService.EditMessage(sessionName, chatID, messageID, req.Text)
	if err != nil {
		log.Printf("[WHATSAPP] EditMessage - Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao editar mensagem"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mensagem editada com sucesso"})
}

// DeleteMessage deleta uma mensagem
func (h *WhatsAppMessageHandler) DeleteMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sessionName := fmt.Sprintf("user_%s", userID)
	chatID := c.Param("chatId")
	messageID := c.Param("messageId")

	log.Printf("[WHATSAPP] DeleteMessage - SessionName: %s, ChatID: %s, MessageID: %s", sessionName, chatID, messageID)

	_, err := h.whatsappService.DeleteMessage(sessionName, chatID, messageID)
	if err != nil {
		log.Printf("[WHATSAPP] DeleteMessage - Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar mensagem"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mensagem deletada com sucesso"})
}

// StarMessage favorita/desfavorita uma mensagem
func (h *WhatsAppMessageHandler) StarMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sessionName := fmt.Sprintf("user_%s", userID)

	var req struct {
		MessageID string `json:"messageId" binding:"required"`
		Star      bool   `json:"star"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[WHATSAPP] StarMessage - SessionName: %s, MessageID: %s, Star: %v", sessionName, req.MessageID, req.Star)

	_, err := h.whatsappService.StarMessage(sessionName, req.MessageID, req.Star)
	if err != nil {
		log.Printf("[WHATSAPP] StarMessage - Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao favoritar mensagem"})
		return
	}

	action := "favoritada"
	if !req.Star {
		action = "desfavoritada"
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Mensagem %s com sucesso", action)})
}

// AddReaction adiciona reação a uma mensagem
func (h *WhatsAppMessageHandler) AddReaction(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sessionName := fmt.Sprintf("user_%s", userID)

	var req struct {
		MessageID string `json:"messageId" binding:"required"`
		Reaction  string `json:"reaction" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[WHATSAPP] AddReaction - SessionName: %s, MessageID: %s, Reaction: %s", sessionName, req.MessageID, req.Reaction)

	err := h.whatsappService.AddReaction(sessionName, req.MessageID, req.Reaction)
	if err != nil {
		log.Printf("[WHATSAPP] AddReaction - Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao adicionar reação"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reação adicionada com sucesso"})
}

// RemoveReaction remove reação de uma mensagem
func (h *WhatsAppMessageHandler) RemoveReaction(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sessionName := fmt.Sprintf("user_%s", userID)

	var req struct {
		MessageID string `json:"messageId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[WHATSAPP] RemoveReaction - SessionName: %s, MessageID: %s", sessionName, req.MessageID)

	err := h.whatsappService.RemoveReaction(sessionName, req.MessageID)
	if err != nil {
		log.Printf("[WHATSAPP] RemoveReaction - Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover reação"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reação removida com sucesso"})
}
