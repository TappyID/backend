package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"tappyone/internal/services"
	"tappyone/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// WhatsAppMediaHandler gerencia envio de mídia no WhatsApp
type WhatsAppMediaHandler struct {
	whatsappService *services.WhatsAppService
	authService     *services.AuthService
}

func NewWhatsAppMediaHandler(whatsappService *services.WhatsAppService, authService *services.AuthService) *WhatsAppMediaHandler {
	return &WhatsAppMediaHandler{
		whatsappService: whatsappService,
		authService:     authService,
	}
}

// GetChats obtém lista de chats
func (h *WhatsAppMediaHandler) GetChats(c *gin.Context) {
	sessionName := "user_" + c.GetString("user_id")
	
	chats, err := h.whatsappService.GetChats(sessionName)
	if err != nil {
		log.Printf("[HANDLER] GetChats error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, chats)
}

// GetContacts obtém lista de contatos
func (h *WhatsAppMediaHandler) GetContacts(c *gin.Context) {
	sessionName := "user_" + c.GetString("user_id")
	
	contacts, err := h.whatsappService.GetContacts(sessionName)
	if err != nil {
		log.Printf("[HANDLER] GetContacts error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, contacts)
}

// GetChatMessages obtém mensagens de um chat
func (h *WhatsAppMediaHandler) GetChatMessages(c *gin.Context) {
	sessionName := "user_" + c.GetString("user_id")
	chatID := c.Param("chatId")
	
	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}
	
	messages, err := h.whatsappService.GetChatMessages(sessionName, chatID, limit, offset)
	if err != nil {
		log.Printf("[HANDLER] GetChatMessages error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, messages)
}

// SendMessage envia mensagem de texto
func (h *WhatsAppMediaHandler) SendMessage(c *gin.Context) {
	sessionName := "user_" + c.GetString("user_id")
	chatID := c.Param("chatId")
	
	var req struct {
		Text string `json:"text" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	result, err := h.whatsappService.SendMessage(sessionName, chatID, req.Text)
	if err != nil {
		log.Printf("[HANDLER] SendMessage error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// UploadFile faz upload de arquivo e retorna URL
func (h *WhatsAppMediaHandler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não encontrado"})
		return
	}
	defer file.Close()

	// Criar diretório de uploads se não existir
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar diretório"})
		return
	}

	// Gerar nome único para o arquivo
	filename := uuid.New().String() + filepath.Ext(header.Filename)
	filePath := filepath.Join(uploadDir, filename)

	// Salvar arquivo
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar arquivo"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao copiar arquivo"})
		return
	}

	// Retornar URL do arquivo
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		// Para Docker, usar host.docker.internal para acessar o host
		baseURL = "http://host.docker.internal:8080"
	}
	
	fileURL := baseURL + "/uploads/" + filename
	
	c.JSON(http.StatusOK, gin.H{
		"url":      fileURL,
		"filename": header.Filename,
		"size":     header.Size,
		"type":     header.Header.Get("Content-Type"),
	})
}

// SendImage envia imagem
func (h *WhatsAppMediaHandler) SendImage(c *gin.Context) {
	sessionName := "user_" + c.GetString("user_id")
	chatID := c.Param("chatId")
	
	var req struct {
		ImageURL string `json:"imageUrl" binding:"required"`
		Caption  string `json:"caption"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	result, err := h.whatsappService.SendImage(sessionName, chatID, req.ImageURL, req.Caption)
	if err != nil {
		log.Printf("[HANDLER] SendImage error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// SendFile envia arquivo
func (h *WhatsAppMediaHandler) SendFile(c *gin.Context) {
	sessionName := "user_" + c.GetString("user_id")
	chatID := c.Param("chatId")
	
	var req struct {
		FileURL  string `json:"fileUrl" binding:"required"`
		Filename string `json:"filename" binding:"required"`
		Caption  string `json:"caption"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	result, err := h.whatsappService.SendFile(sessionName, chatID, req.FileURL, req.Filename, req.Caption)
	if err != nil {
		log.Printf("[HANDLER] SendFile error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// SendVoice envia áudio
func (h *WhatsAppMediaHandler) SendVoice(c *gin.Context) {
	sessionName := "user_" + c.GetString("user_id")
	chatID := c.Param("chatId")
	
	var req struct {
		AudioURL string `json:"audioUrl" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	result, err := h.whatsappService.SendVoice(sessionName, chatID, req.AudioURL)
	if err != nil {
		log.Printf("[HANDLER] SendVoice error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// SendVideo envia vídeo
func (h *WhatsAppMediaHandler) SendVideo(c *gin.Context) {
	sessionName := "user_" + c.GetString("user_id")
	chatID := c.Param("chatId")
	
	var req struct {
		VideoURL string `json:"videoUrl" binding:"required"`
		Caption  string `json:"caption"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	result, err := h.whatsappService.SendVideo(sessionName, chatID, req.VideoURL, req.Caption)
	if err != nil {
		log.Printf("[HANDLER] SendVideo error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// GetPresence obtém status de presença
func (h *WhatsAppMediaHandler) GetPresence(c *gin.Context) {
	sessionName := "user_" + c.GetString("user_id")
	
	presence, err := h.whatsappService.GetPresence(sessionName)
	if err != nil {
		log.Printf("[HANDLER] GetPresence error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, presence)
}

// MarkAsRead marca mensagens como lidas
func (h *WhatsAppMediaHandler) MarkAsRead(c *gin.Context) {
	sessionName := "user_" + c.GetString("user_id")
	chatID := c.Param("chatId")
	
	result, err := h.whatsappService.MarkAsRead(sessionName, chatID)
	if err != nil {
		log.Printf("[HANDLER] MarkAsRead error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// SetTyping define status de digitação
func (h *WhatsAppMediaHandler) SetTyping(c *gin.Context) {
	sessionName := "user_" + c.GetString("user_id")
	chatID := c.Param("chatId")
	
	var req struct {
		Presence string `json:"presence" binding:"required"` // "typing" ou "paused"
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	result, err := h.whatsappService.SetTyping(sessionName, chatID, req.Presence)
	if err != nil {
		log.Printf("[HANDLER] SetTyping error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// UploadAndSendMedia faz upload e envia mídia em uma única operação
func (h *WhatsAppMediaHandler) UploadAndSendMedia(c *gin.Context) {
	// Validação manual do JWT
	userID, err := utils.ValidateJWTFromHeader(c, h.authService)
	if err != nil {
		log.Printf("[HANDLER] UploadAndSendMedia - Auth failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	sessionName := "user_" + userID
	chatID := c.Param("chatId")
	mediaType := c.PostForm("type") // "image", "file", "voice", "video"
	caption := c.PostForm("caption")
	
	// Upload do arquivo
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arquivo não encontrado"})
		return
	}
	defer file.Close()

	// Criar diretório de uploads se não existir
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar diretório"})
		return
	}

	// Gerar nome único para o arquivo
	filename := uuid.New().String() + filepath.Ext(header.Filename)
	filePath := filepath.Join(uploadDir, filename)

	// Salvar arquivo
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar arquivo"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao copiar arquivo"})
		return
	}

	// Gerar URL do arquivo
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		// Para Docker, usar host.docker.internal para acessar o host
		baseURL = "http://host.docker.internal:8080"
	}
	fileURL := baseURL + "/uploads/" + filename

	// Enviar mídia baseado no tipo
	switch mediaType {
	case "image":
		_, err = h.whatsappService.SendImage(sessionName, chatID, fileURL, caption)
	case "file":
		_, err = h.whatsappService.SendFile(sessionName, chatID, fileURL, header.Filename, caption)
	case "voice":
		// Para áudio, usar URL como os outros tipos
		_, err = h.whatsappService.SendVoice(sessionName, chatID, fileURL)
	case "video":
		_, err = h.whatsappService.SendVideo(sessionName, chatID, fileURL, caption)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de mídia inválido"})
		return
	}

	if err != nil {
		log.Printf("[HANDLER] UploadAndSendMedia error: %v", err)
		// Remover arquivo se falhou o envio
		os.Remove(filePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":  "File sent successfully",
		"fileUrl":  fileURL,
		"filename": header.Filename,
		"size":     header.Size,
		"type":     mediaType,
	})
}

// SendVideoMessage envia vídeo via WhatsApp
func (h *WhatsAppMediaHandler) SendVideoMessage(c *gin.Context) {
	// Validar token JWT
	userID, err := utils.ValidateJWTFromHeader(c, h.authService)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	sessionName := "user_" + userID
	chatID := c.Param("chatId")

	// Parse multipart form
	err = c.Request.ParseMultipartForm(50 << 20) // 50MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No video file provided"})
		return
	}
	defer file.Close()

	// Ler arquivo
	videoData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read video file"})
		return
	}

	// Obter caption opcional
	caption := c.Request.FormValue("caption")

	// Enviar vídeo
	err = h.whatsappService.SendVideoMessage(sessionName, chatID, videoData, header.Filename, caption)
	if err != nil {
		log.Printf("[HANDLER] SendVideoMessage error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Video sent successfully"})
}

// Função auxiliar para validar tipos de arquivo
func isValidFileType(filename, allowedTypes string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	types := strings.Split(allowedTypes, ",")
	
	for _, t := range types {
		if ext == strings.TrimSpace(t) {
			return true
		}
	}
	return false
}

// Função auxiliar para obter tipo de mídia baseado na extensão
func getMediaTypeFromExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv"}
	audioExts := []string{".mp3", ".wav", ".ogg", ".m4a", ".aac"}
	
	for _, e := range imageExts {
		if ext == e {
			return "image"
		}
	}
	
	for _, e := range videoExts {
		if ext == e {
			return "video"
		}
	}
	
	for _, e := range audioExts {
		if ext == e {
			return "voice"
		}
	}
	
	return "file"
}

// ForwardMessage encaminha uma mensagem para outro chat
func (h *WhatsAppMediaHandler) ForwardMessage(c *gin.Context) {
	// Validar JWT
	userID, err := utils.ValidateJWTFromHeader(c, h.authService)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	var req struct {
		ToChatID  string `json:"toChatId" binding:"required"`
		MessageID string `json:"messageId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get session name from user
	sessionName := fmt.Sprintf("user_%s", userID)
	log.Printf("[HANDLER] ForwardMessage - userID: %s, sessionName: %s, toChatID: %s, messageID: %s", userID, sessionName, req.ToChatID, req.MessageID)

	// Encaminhar mensagem via WhatsApp Service
	result, err := h.whatsappService.ForwardMessage(sessionName, req.ToChatID, req.MessageID)
	if err != nil {
		log.Printf("[HANDLER] ForwardMessage - WhatsApp service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao encaminhar mensagem"})
		return
	}

	log.Printf("[HANDLER] ForwardMessage - Success!")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// EditMessage edita uma mensagem existente
func (h *WhatsAppMediaHandler) EditMessage(c *gin.Context) {
	// Validar JWT
	userID, err := utils.ValidateJWTFromHeader(c, h.authService)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	chatID := c.Param("chatId")
	messageID := c.Param("messageId")

	var req struct {
		Text string `json:"text" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get session name from user
	sessionName := fmt.Sprintf("user_%s", userID)
	log.Printf("[HANDLER] EditMessage - userID: %s, sessionName: %s, chatID: %s, messageID: %s", userID, sessionName, chatID, messageID)

	// Editar mensagem via WhatsApp Service
	result, err := h.whatsappService.EditMessage(sessionName, chatID, messageID, req.Text)
	if err != nil {
		log.Printf("[HANDLER] EditMessage - WhatsApp service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao editar mensagem"})
		return
	}

	log.Printf("[HANDLER] EditMessage - Success!")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// DeleteMessage deleta uma mensagem
func (h *WhatsAppMediaHandler) DeleteMessage(c *gin.Context) {
	// Validar JWT
	userID, err := utils.ValidateJWTFromHeader(c, h.authService)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	chatID := c.Param("chatId")
	messageID := c.Param("messageId")

	// Get session name from user
	sessionName := fmt.Sprintf("user_%s", userID)
	log.Printf("[HANDLER] DeleteMessage - userID: %s, sessionName: %s, chatID: %s, messageID: %s", userID, sessionName, chatID, messageID)

	// Deletar mensagem via WhatsApp Service
	result, err := h.whatsappService.DeleteMessage(sessionName, chatID, messageID)
	if err != nil {
		log.Printf("[HANDLER] DeleteMessage - WhatsApp service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar mensagem"})
		return
	}

	log.Printf("[HANDLER] DeleteMessage - Success!")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// StarMessage favorita/desfavorita uma mensagem
func (h *WhatsAppMediaHandler) StarMessage(c *gin.Context) {
	// Validar JWT
	userID, err := utils.ValidateJWTFromHeader(c, h.authService)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	var req struct {
		MessageID string `json:"messageId" binding:"required"`
		Star      bool   `json:"star"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get session name from user
	sessionName := fmt.Sprintf("user_%s", userID)
	log.Printf("[HANDLER] StarMessage - userID: %s, sessionName: %s, messageID: %s, star: %t", userID, sessionName, req.MessageID, req.Star)

	// Favoritar mensagem via WhatsApp Service
	result, err := h.whatsappService.StarMessage(sessionName, req.MessageID, req.Star)
	if err != nil {
		log.Printf("[HANDLER] StarMessage - WhatsApp service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao favoritar mensagem"})
		return
	}

	log.Printf("[HANDLER] StarMessage - Success!")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// SendContactVcard envia um contato via vCard
func (h *WhatsAppMediaHandler) SendContactVcard(c *gin.Context) {
	// Validar JWT
	userID, err := utils.ValidateJWTFromHeader(c, h.authService)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	var req struct {
		ChatID    string `json:"chatId" binding:"required"`
		ContactID string `json:"contactId" binding:"required"`
		Name      string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get session name from user
	sessionName := fmt.Sprintf("user_%s", userID)
	log.Printf("[HANDLER] SendContactVcard - userID: %s, sessionName: %s, chatID: %s", userID, sessionName, req.ChatID)

	// Enviar contato via WhatsApp Service
	result, err := h.whatsappService.SendContactVcard(sessionName, req.ChatID, req.ContactID, req.Name)
	if err != nil {
		log.Printf("[HANDLER] SendContactVcard - WhatsApp service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar contato"})
		return
	}

	log.Printf("[HANDLER] SendContactVcard - Success!")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// SendLocation envia uma localização
func (h *WhatsAppMediaHandler) SendLocation(c *gin.Context) {
	// Validar JWT
	userID, err := utils.ValidateJWTFromHeader(c, h.authService)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	var req struct {
		ChatID    string  `json:"chatId" binding:"required"`
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
		Title     string  `json:"title"`
		Address   string  `json:"address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get session name from user
	sessionName := fmt.Sprintf("user_%s", userID)
	log.Printf("[HANDLER] SendLocation - userID: %s, sessionName: %s, chatID: %s", userID, sessionName, req.ChatID)

	// Enviar localização via WhatsApp Service
	result, err := h.whatsappService.SendLocation(sessionName, req.ChatID, req.Latitude, req.Longitude, req.Title, req.Address)
	if err != nil {
		log.Printf("[HANDLER] SendLocation - WhatsApp service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar localização"})
		return
	}

	log.Printf("[HANDLER] SendLocation - Success!")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// SendPoll envia uma enquete
func (h *WhatsAppMediaHandler) SendPoll(c *gin.Context) {
	// Validar JWT
	userID, err := utils.ValidateJWTFromHeader(c, h.authService)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	var req struct {
		ChatID          string   `json:"chatId" binding:"required"`
		Name            string   `json:"name" binding:"required"`
		Options         []string `json:"options" binding:"required"`
		MultipleAnswers bool     `json:"multipleAnswers"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get session name from user
	sessionName := fmt.Sprintf("user_%s", userID)
	log.Printf("[HANDLER] SendPoll - userID: %s, sessionName: %s, chatID: %s", userID, sessionName, req.ChatID)

	// Enviar enquete via WhatsApp Service
	result, err := h.whatsappService.SendPoll(sessionName, req.ChatID, req.Name, req.Options, req.MultipleAnswers)
	if err != nil {
		log.Printf("[HANDLER] SendPoll - WhatsApp service error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar enquete"})
		return
	}

	log.Printf("[HANDLER] SendPoll - Success!")
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
