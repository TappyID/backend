package handlers

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	
	messages, err := h.whatsappService.GetChatMessages(sessionName, chatID)
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
	var result interface{}
	switch mediaType {
	case "image":
		result, err = h.whatsappService.SendImage(sessionName, chatID, fileURL, caption)
	case "file":
		result, err = h.whatsappService.SendFile(sessionName, chatID, fileURL, header.Filename, caption)
	case "voice":
		// Para áudio, usar URL como os outros tipos
		result, err = h.whatsappService.SendVoice(sessionName, chatID, fileURL)
	case "video":
		result, err = h.whatsappService.SendVideo(sessionName, chatID, fileURL, caption)
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
		"result":   result,
		"fileUrl":  fileURL,
		"filename": header.Filename,
		"size":     header.Size,
		"type":     mediaType,
	})
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
