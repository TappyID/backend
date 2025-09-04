package handlers

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"tappyone/internal/models"
	"tappyone/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthHandler gerencia autenticação
type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.Login(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := h.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	// Limpar senha antes de retornar
	user.Senha = ""
	c.JSON(http.StatusOK, user)
}

// UserHandler gerencia usuários
type UserHandler struct {
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetString("user_id")
	user, err := h.userService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	// TODO: Implementar atualização do usuário
	c.JSON(http.StatusOK, gin.H{"message": "Atualização do usuário"})
}

func (h *UserHandler) List(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Filtros opcionais
	tipo := c.Query("tipo")
	status := c.Query("status")
	search := c.Query("search")

	users, err := h.userService.ListUsers(userID, tipo, status, search)
	if err != nil {
		log.Printf("Erro ao listar usuários: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar usuários"})
		return
	}

	// Limpar senhas antes de retornar
	for i := range users {
		users[i].Senha = ""
	}

	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Nome     string             `json:"nome" binding:"required"`
		Email    string             `json:"email" binding:"required,email"`
		Telefone *string            `json:"telefone"`
		Tipo     models.TipoUsuario `json:"tipo" binding:"required"`
		Senha    string             `json:"senha" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar se email já existe
	existingUser, _ := h.userService.GetByEmail(req.Email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email já está em uso"})
		return
	}

	// Hash da senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Senha), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Erro ao gerar hash da senha: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao processar senha"})
		return
	}

	usuario := &models.Usuario{
		Nome:     req.Nome,
		Email:    req.Email,
		Telefone: req.Telefone,
		Tipo:     req.Tipo,
		Senha:    string(hashedPassword),
		Ativo:    true,
	}

	if err := h.userService.CreateUserWithDefaults(usuario); err != nil {
		log.Printf("Erro ao criar usuário: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar usuário"})
		return
	}

	// Limpar senha antes de retornar
	usuario.Senha = ""
	c.JSON(http.StatusCreated, usuario)
}

func (h *UserHandler) GetByID(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID é obrigatório"})
		return
	}

	usuario, err := h.userService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	// Buscar estatísticas do usuário
	stats, _ := h.userService.GetUserStats(userID)
	usuario.Senha = "" // Limpar senha

	response := map[string]interface{}{
		"usuario":      usuario,
		"estatisticas": stats,
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) Update(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID é obrigatório"})
		return
	}

	var req struct {
		Nome     *string             `json:"nome"`
		Email    *string             `json:"email"`
		Telefone *string             `json:"telefone"`
		Tipo     *models.TipoUsuario `json:"tipo"`
		Ativo    *bool               `json:"ativo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Buscar usuário existente
	usuario, err := h.userService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	// Atualizar campos se fornecidos
	if req.Nome != nil {
		usuario.Nome = *req.Nome
	}
	if req.Email != nil {
		// Verificar se email já existe (excluindo o próprio usuário)
		existingUser, _ := h.userService.GetByEmail(*req.Email)
		if existingUser != nil && existingUser.ID != userID {
			c.JSON(http.StatusConflict, gin.H{"error": "Email já está em uso"})
			return
		}
		usuario.Email = *req.Email
	}
	if req.Telefone != nil {
		usuario.Telefone = req.Telefone
	}
	if req.Tipo != nil {
		usuario.Tipo = *req.Tipo
	}
	if req.Ativo != nil {
		usuario.Ativo = *req.Ativo
	}

	if err := h.userService.Update(usuario); err != nil {
		log.Printf("Erro ao atualizar usuário: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar usuário"})
		return
	}

	// Limpar senha antes de retornar
	usuario.Senha = ""
	c.JSON(http.StatusOK, usuario)
}

func (h *UserHandler) Delete(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID é obrigatório"})
		return
	}

	// Verificar se usuário existe
	_, err := h.userService.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Usuário não encontrado"})
		return
	}

	// Soft delete - apenas desativar usuário
	if err := h.userService.DeleteUser(userID); err != nil {
		log.Printf("Erro ao desativar usuário: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao desativar usuário"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Usuário desativado com sucesso"})
}

// WhatsAppHandler gerencia WhatsApp
type WhatsAppHandler struct {
	whatsappService *services.WhatsAppService
}

func NewWhatsAppHandler(whatsappService *services.WhatsAppService) *WhatsAppHandler {
	return &WhatsAppHandler{whatsappService: whatsappService}
}

func (h *WhatsAppHandler) CreateSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		NomeSessao string `json:"nomeSessao" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[WHATSAPP] Error parsing request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[WHATSAPP] Creating session %s for user %s", req.NomeSessao, userID)

	// Criar sessão no banco
	session := &models.SessaoWhatsApp{
		NomeSessao: req.NomeSessao,
		Status:     models.StatusSessaoDesconectado,
		Ativo:      true,
		UsuarioID:  userID.(string),
	}

	if err := h.whatsappService.CreateSession(session); err != nil {
		log.Printf("[WHATSAPP] Error creating session in DB: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	log.Printf("[WHATSAPP] Session created successfully with ID: %s", session.ID)
	c.JSON(http.StatusCreated, session)
}

func (h *WhatsAppHandler) ListSessions(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var sessoes []models.SessaoWhatsApp

	err := h.whatsappService.GetDB().Where("usuario_id = ? AND ativo = ?", userID, true).Find(&sessoes).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar sessões"})
		return
	}

	c.JSON(http.StatusOK, sessoes)
}

func (h *WhatsAppHandler) GetSession(c *gin.Context) {
	// TODO: Implementar busca de sessão
	c.JSON(http.StatusOK, gin.H{"message": "Busca de sessão WhatsApp"})
}

func (h *WhatsAppHandler) StartSession(c *gin.Context) {
	// TODO: Implementar início de sessão
	c.JSON(http.StatusOK, gin.H{"message": "Início de sessão WhatsApp"})
}

func (h *WhatsAppHandler) StopSession(c *gin.Context) {
	// TODO: Implementar parada de sessão
	c.JSON(http.StatusOK, gin.H{"message": "Parada de sessão WhatsApp"})
}

func (h *WhatsAppHandler) GetQRCode(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
		return
	}

	// Obter configuração do WAHA
	wahaURL := os.Getenv("WAHA_URL")
	if wahaURL == "" {
		wahaURL = "http://159.65.34.199:3001"
	}
	wahaAPIKey := os.Getenv("WAHA_API_KEY")
	if wahaAPIKey == "" {
		wahaAPIKey = "tappyone-waha-2024-secretkey"
	}

	// Tentar diferentes endpoints para obter QR code
	endpoints := []string{
		fmt.Sprintf("%s/api/sessions/%s/auth/qr?format=image", wahaURL, sessionID),
		fmt.Sprintf("%s/api/sessions/%s/qr?format=image", wahaURL, sessionID),
		fmt.Sprintf("%s/api/sessions/%s/qr", wahaURL, sessionID),
		fmt.Sprintf("%s/api/sessions/%s/screenshot", wahaURL, sessionID),
	}

	for _, endpoint := range endpoints {
		log.Printf("[QR] Tentando endpoint: %s", endpoint)
		
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			log.Printf("[QR] Erro ao criar request: %v", err)
			continue
		}

		req.Header.Set("X-Api-Key", wahaAPIKey)
		req.Header.Set("Accept", "image/png")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("[QR] Erro ao fazer request: %v", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			// QR code encontrado, retornar como imagem
			c.Header("Content-Type", "image/png")
			c.Header("Cache-Control", "no-cache")
			
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("[QR] Erro ao ler response body: %v", err)
				continue
			}

			if len(body) > 0 {
				log.Printf("[QR] QR code obtido via: %s", endpoint)
				c.Data(http.StatusOK, "image/png", body)
				return
			}
		} else {
			log.Printf("[QR] Endpoint %s retornou: %d", endpoint, resp.StatusCode)
		}
	}

	// Se chegou aqui, nenhum endpoint funcionou
	log.Printf("[QR] Nenhum endpoint de QR code funcionou para sessão: %s", sessionID)
	c.JSON(http.StatusNotFound, gin.H{
		"error": "QR Code não encontrado. Verifique se a sessão está no estado SCAN_QR_CODE",
		"session": sessionID,
		"available_in_logs": "docker logs backend-waha-1",
	})
}

func (h *WhatsAppHandler) WebhookHandler(c *gin.Context) {
	var webhookData map[string]interface{}
	if err := c.ShouldBindJSON(&webhookData); err != nil {
		log.Printf("Erro ao fazer bind do webhook: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	log.Printf("Webhook recebido: %+v", webhookData)

	// Processar evento de mudança de status da sessão
	if event, ok := webhookData["event"].(string); ok && event == "session.status" {
		go h.processSessionStatusChange(webhookData)
	}

	// Verificar se é uma mensagem de mídia
	if event, ok := webhookData["event"].(string); ok && event == "message" {
		if data, ok := webhookData["data"].(map[string]interface{}); ok {
			if msgType, ok := data["type"].(string); ok {
				// Processar mensagens de mídia
				if msgType == "image" || msgType == "video" || msgType == "audio" || msgType == "voice" || msgType == "document" {
					go h.processMediaMessage(data)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processado com sucesso"})
}

func (h *WhatsAppHandler) processSessionStatusChange(webhookData map[string]interface{}) {
	// Extrair dados do webhook
	sessionName, ok := webhookData["session"].(string)
	if !ok {
		log.Printf("Erro: nome da sessão não encontrado no webhook")
		return
	}

	data, ok := webhookData["data"].(map[string]interface{})
	if !ok {
		log.Printf("Erro: dados inválidos no webhook session.status")
		return
	}

	newStatus, ok := data["status"].(string)
	if !ok {
		log.Printf("Erro: status não encontrado nos dados do webhook")
		return
	}

	log.Printf("Mudança de status detectada - Sessão: %s, Novo Status: %s", sessionName, newStatus)

	// Extrair user_id do nome da sessão (formato: user_{uuid})
	if !strings.HasPrefix(sessionName, "user_") {
		log.Printf("Erro: formato de sessão inválido: %s", sessionName)
		return
	}

	userID := strings.TrimPrefix(sessionName, "user_")
	log.Printf("Extraído user_id: %s", userID)

	// Buscar conexão no banco
	db := h.whatsappService.GetDB()
	var connection models.UserConnection
	result := db.Where("user_id = ? AND platform = ?", userID, "whatsapp").First(&connection)

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
		if err := db.Save(&connection).Error; err != nil {
			log.Printf("Erro ao atualizar status da conexão: %v", err)
			return
		}

		log.Printf("Status da conexão WhatsApp atualizado com sucesso para: %s", ourStatus)
	} else {
		log.Printf("Status já está atualizado: %s", ourStatus)
	}
}

func (h *WhatsAppHandler) processMediaMessage(data map[string]interface{}) {
	// Extrair informações da mensagem
	chatID, _ := data["from"].(string)
	messageID, _ := data["id"].(string)
	msgType, _ := data["type"].(string)

	log.Printf("Processando mídia: chatID=%s, messageID=%s, type=%s", chatID, messageID, msgType)

	// Verificar se há mídia com hasMedia e media.url
	hasMedia, _ := data["hasMedia"].(bool)
	var mediaURL string
	var filename string

	if hasMedia {
		if media, ok := data["media"].(map[string]interface{}); ok {
			if url, ok := media["url"].(string); ok {
				mediaURL = url
			}
			if fname, ok := media["filename"].(string); ok {
				filename = fname
			}
		}
	}

	if mediaURL != "" {
		log.Printf("Baixando mídia de: %s", mediaURL)

		mediaData, err := h.downloadMediaFromURL(mediaURL)
		if err != nil {
			log.Printf("Erro ao fazer download da mídia %s: %v", mediaURL, err)
			return
		}

		// Salvar mídia no droplet
		savedURL, err := saveMediaToDroplet(mediaData, filename)
		if err != nil {
			log.Printf("Erro ao salvar mídia no droplet: %v", err)
			return
		}

		log.Printf("Mídia salva no droplet: %s", savedURL)
		log.Printf("Mídia recebida - Chat: %s, MessageID: %s, URL: %s, Tipo: %s", chatID, messageID, savedURL, msgType)
	}
}

func (h *WhatsAppHandler) downloadMediaFromURL(mediaURL string) ([]byte, error) {
	// Fazer requisição HTTP para baixar a mídia
	resp, err := http.Get(mediaURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao fazer requisição: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code inválido: %d", resp.StatusCode)
	}

	// Ler dados da resposta
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler dados: %v", err)
	}

	return data, nil
}

func (h *WhatsAppHandler) saveMediaToDroplet(data []byte, filename, msgType string) (string, error) {
	// Criar diretório uploads se não existir
	uploadsDir := "./uploads"
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		return "", fmt.Errorf("erro ao criar diretório uploads: %v", err)
	}

	// Gerar nome único para o arquivo
	timestamp := time.Now().Unix()
	randomId := fmt.Sprintf("%d", rand.Int63())

	// Determinar extensão baseada no tipo
	var ext string
	switch msgType {
	case "image":
		ext = ".jpg"
		if strings.Contains(filename, ".png") {
			ext = ".png"
		} else if strings.Contains(filename, ".gif") {
			ext = ".gif"
		} else if strings.Contains(filename, ".webp") {
			ext = ".webp"
		}
	case "audio", "voice":
		ext = ".oga"
		if strings.Contains(filename, ".mp3") {
			ext = ".mp3"
		} else if strings.Contains(filename, ".wav") {
			ext = ".wav"
		}
	case "video":
		ext = ".mp4"
		if strings.Contains(filename, ".webm") {
			ext = ".webm"
		}
	case "document":
		// Manter extensão original
		if idx := strings.LastIndex(filename, "."); idx != -1 {
			ext = filename[idx:]
		} else {
			ext = ".bin"
		}
	default:
		ext = ".bin"
	}

	// Nome final do arquivo
	finalFilename := fmt.Sprintf("%s_%d_%s%s", msgType, timestamp, randomId, ext)
	filePath := filepath.Join(uploadsDir, finalFilename)

	// Salvar arquivo
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("erro ao salvar arquivo: %v", err)
	}

	// Retornar URL pública
	publicURL := fmt.Sprintf("http://159.65.34.199:3001/api/files/%s", finalFilename)
	log.Printf("Mídia salva no droplet: %s", publicURL)

	return publicURL, nil
}

// SendImageMessage handler para envio de imagens
func (h *WhatsAppHandler) SendImageMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sessionName := fmt.Sprintf("user_%s", userID)
	chatID := c.Param("chatId")

	// Parse multipart form
	err := c.Request.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar form"})
		return
	}

	// Get image file
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhuma imagem fornecida"})
		return
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler arquivo"})
		return
	}

	// Save to droplet
	mediaURL, err := h.saveMediaToDroplet(fileData, header.Filename, "image")
	if err != nil {
		log.Printf("Erro ao salvar imagem no droplet: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar imagem"})
		return
	}

	// Get caption
	caption := c.Request.FormValue("caption")

	// Send via WAHA API
	err = h.whatsappService.SendImageMessage(sessionName, chatID, fileData, header.Filename, caption)
	if err != nil {
		log.Printf("Erro ao enviar imagem via WAHA: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar imagem"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"mediaUrl": mediaURL,
		"message":  "Imagem enviada com sucesso",
	})
}

// SendVoiceMessage handler para envio de áudios
func (h *WhatsAppHandler) SendVoiceMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sessionName := fmt.Sprintf("user_%s", userID)
	chatID := c.Param("chatId")

	// Parse multipart form
	err := c.Request.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar form"})
		return
	}

	// Get audio file
	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhum áudio fornecido"})
		return
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler arquivo"})
		return
	}

	// Save to droplet
	mediaURL, err := saveMediaToDroplet(fileData, header.Filename)
	if err != nil {
		log.Printf("Erro ao salvar áudio no droplet: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar áudio"})
		return
	}

	// Send via WAHA API with convert=true for compatibility
	err = h.whatsappService.SendVoiceMessage(sessionName, chatID, fileData, header.Filename)
	if err != nil {
		log.Printf("Erro ao enviar áudio via WAHA: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar áudio"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"mediaUrl": mediaURL,
		"message":  "Áudio enviado com sucesso",
	})
}

// SendFileMessage handler para envio de documentos
func (h *WhatsAppHandler) SendFileMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sessionName := fmt.Sprintf("user_%s", userID)
	chatID := c.Param("chatId")

	// Parse multipart form
	err := c.Request.ParseMultipartForm(32 << 20) // 32MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erro ao processar form"})
		return
	}

	// Get file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nenhum arquivo fornecido"})
		return
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao ler arquivo"})
		return
	}

	// Save to droplet
	mediaURL, err := saveMediaToDroplet(fileData, header.Filename)
	if err != nil {
		log.Printf("Erro ao salvar arquivo no droplet: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar arquivo"})
		return
	}

	// Get caption if provided
	caption := c.PostForm("caption")

	// Send via WAHA API
	err = h.whatsappService.SendFileMessage(sessionName, chatID, fileData, header.Filename, caption)
	if err != nil {
		log.Printf("Erro ao enviar arquivo via WAHA: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao enviar arquivo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"mediaUrl": mediaURL,
		"message":  "Arquivo enviado com sucesso",
	})
}

// DownloadMedia handler para download de mídia
func (h *WhatsAppHandler) DownloadMedia(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sessionName := fmt.Sprintf("user_%s", userID)
	mediaID := c.Param("mediaId")

	if mediaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mediaId é obrigatório"})
		return
	}

	// Download via WAHA API
	mediaData, filename, err := h.whatsappService.DownloadMedia(sessionName, mediaID)
	if err != nil {
		log.Printf("[WHATSAPP] DownloadMedia - Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao baixar mídia"})
		return
	}

	// Return file
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, "application/octet-stream", mediaData)
}

// saveMediaToDroplet salva mídia no droplet e retorna URL pública
func saveMediaToDroplet(data []byte, filename string) (string, error) {
	// Criar diretório se não existir
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", err
	}

	// Gerar nome único para o arquivo
	ext := filepath.Ext(filename)
	uniqueFilename := fmt.Sprintf("%s-%d%s",
		strings.TrimSuffix(filename, ext),
		time.Now().Unix()+rand.Int63n(1000),
		ext)

	filePath := filepath.Join(uploadDir, uniqueFilename)

	// Salvar arquivo
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", err
	}

	// Retornar URL pública para acessar o arquivo
	return fmt.Sprintf("http://159.65.34.199:3001/api/files/%s", uniqueFilename), nil
}

// AtendimentoStatsHandler gerencia estatísticas de atendimento
type AtendimentoStatsHandler struct {
	whatsappService *services.WhatsAppService
	db              *gorm.DB
}

func NewAtendimentoStatsHandler(whatsappService *services.WhatsAppService, db *gorm.DB) *AtendimentoStatsHandler {
	return &AtendimentoStatsHandler{
		whatsappService: whatsappService,
		db:              db,
	}
}

type AtendimentoStats struct {
	ChatsAtivos        int `json:"chats_ativos"`
	AtendentesOnline   int `json:"atendentes_online"`
	MensagensPendentes int `json:"mensagens_pendentes"`
	TempoRespostaMedio int `json:"tempo_resposta_medio"`
}

func (h *AtendimentoStatsHandler) GetStats(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autorizado"})
		return
	}

	// Buscar chats do WhatsApp
	chatsData, err := h.whatsappService.GetChats(userID)
	chatsAtivos := 0
	totalChats := 0

	if err != nil {
		log.Printf("Erro ao buscar chats: %v", err)
	} else {
		// Processar chats retornados (interface{})
		if chatsSlice, ok := chatsData.([]interface{}); ok {
			totalChats = len(chatsSlice)
			agora := time.Now()
			
			for _, chatInterface := range chatsSlice {
				if chatMap, ok := chatInterface.(map[string]interface{}); ok {
					// Verifica se o timestamp da conversa é das últimas 24h
					if timestampFloat, ok := chatMap["conversationTimestamp"].(float64); ok {
						timestamp := int64(timestampFloat)
						if timestamp > agora.AddDate(0, 0, -1).Unix() {
							chatsAtivos++
						}
					}
				}
			}
		}
	}

	// Contar atendentes online (usuários ativos do sistema)
	var atendentesOnline int64
	h.db.Model(&models.Usuario{}).Where("ativo = ? AND tipo LIKE ?", true, "ATENDENTE%").Count(&atendentesOnline)
	
	// Incluir admins também
	var adminsOnline int64
	h.db.Model(&models.Usuario{}).Where("ativo = ? AND tipo = ?", true, "ADMIN").Count(&adminsOnline)
	
	totalAtendentes := int(atendentesOnline + adminsOnline)

	// Mensagens pendentes (aproximação baseada no total de chats)
	mensagensPendentes := totalChats / 3 // Aproximação: 33% dos chats têm mensagens pendentes

	// Tempo resposta médio em minutos (simulado)
	tempoRespostaMedio := 12

	stats := AtendimentoStats{
		ChatsAtivos:        chatsAtivos,
		AtendentesOnline:   totalAtendentes,
		MensagensPendentes: mensagensPendentes,
		TempoRespostaMedio: tempoRespostaMedio,
	}

	c.JSON(http.StatusOK, stats)
}

// KanbanHandler gerencia Kanban
type KanbanHandler struct {
	kanbanService *services.KanbanService
}

func NewKanbanHandler(kanbanService *services.KanbanService) *KanbanHandler {
	return &KanbanHandler{kanbanService: kanbanService}
}

func (h *KanbanHandler) ListQuadros(c *gin.Context) {
	userID := c.GetString("user_id")
	quadros, err := h.kanbanService.GetQuadrosByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, quadros)
}

func (h *KanbanHandler) CreateQuadro(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		Nome      string  `json:"nome" binding:"required"`
		Cor       string  `json:"cor" binding:"required"`
		Descricao *string `json:"descricao"`
		Posicao   int     `json:"posicao"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quadro := &models.Quadro{
		Nome:      req.Nome,
		Cor:       req.Cor,
		Descricao: req.Descricao,
		Posicao:   req.Posicao,
		UsuarioID: userID,
		Ativo:     true,
	}

	if err := h.kanbanService.CreateQuadro(quadro); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, quadro)
}

func (h *KanbanHandler) GetQuadro(c *gin.Context) {
	quadroID := c.Param("id")
	userID := c.GetString("user_id")

	quadro, err := h.kanbanService.GetQuadroByID(quadroID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Quadro não encontrado"})
		return
	}

	c.JSON(http.StatusOK, quadro)
}

func (h *KanbanHandler) UpdateQuadro(c *gin.Context) {
	quadroID := c.Param("id")
	userID := c.GetString("user_id")

	var req struct {
		Nome      *string `json:"nome"`
		Cor       *string `json:"cor"`
		Descricao *string `json:"descricao"`
		Posicao   *int    `json:"posicao"`
		Ativo     *bool   `json:"ativo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	quadro, err := h.kanbanService.UpdateQuadro(quadroID, userID, req.Nome, req.Cor, req.Descricao, req.Posicao, req.Ativo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, quadro)
}

func (h *KanbanHandler) DeleteQuadro(c *gin.Context) {
	quadroID := c.Param("id")
	userID := c.GetString("user_id")

	if err := h.kanbanService.DeleteQuadro(quadroID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Quadro excluído com sucesso"})
}

// EditColumn edita o nome de uma coluna
func (h *KanbanHandler) EditColumn(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		QuadroID string `json:"quadroId" binding:"required"`
		ColunaID string `json:"colunaId" binding:"required"`
		NovoNome string `json:"novoNome" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[KANBAN] EditColumn - UserID: %s, QuadroID: %s, ColunaID: %s, NovoNome: %s", userID, req.QuadroID, req.ColunaID, req.NovoNome)

	if err := h.kanbanService.EditColumn(req.QuadroID, req.ColunaID, req.NovoNome, userID); err != nil {
		log.Printf("[KANBAN] EditColumn - Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao editar coluna"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Coluna editada com sucesso"})
}

// DeleteColumn exclui uma coluna
func (h *KanbanHandler) DeleteColumn(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		QuadroID string `json:"quadroId" binding:"required"`
		ColunaID string `json:"colunaId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[KANBAN] DeleteColumn - UserID: %s, QuadroID: %s, ColunaID: %s", userID, req.QuadroID, req.ColunaID)

	if err := h.kanbanService.DeleteColumn(req.QuadroID, req.ColunaID, userID); err != nil {
		log.Printf("[KANBAN] DeleteColumn - Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir coluna"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Coluna excluída com sucesso"})
}

// CreateColumn cria uma nova coluna
func (h *KanbanHandler) CreateColumn(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		QuadroID string  `json:"quadroId" binding:"required"`
		Nome     string  `json:"nome" binding:"required"`
		Cor      *string `json:"cor"`
		Posicao  int     `json:"posicao"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[KANBAN] CreateColumn - UserID: %s, QuadroID: %s, Nome: %s, Posicao: %d", userID, req.QuadroID, req.Nome, req.Posicao)

	coluna, err := h.kanbanService.CreateColumn(req.QuadroID, req.Nome, req.Cor, req.Posicao, userID)
	if err != nil {
		log.Printf("[KANBAN] CreateColumn - Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar coluna"})
		return
	}

	c.JSON(http.StatusOK, coluna)
}

// MoveCard move um card entre colunas
func (h *KanbanHandler) MoveCard(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		QuadroID       string `json:"quadroId" binding:"required"`
		CardID         string `json:"cardId" binding:"required"`
		SourceColumnID string `json:"sourceColumnId" binding:"required"`
		TargetColumnID string `json:"targetColumnId" binding:"required"`
		Posicao        int    `json:"posicao"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[KANBAN] MoveCard - UserID: %s, QuadroID: %s, CardID: %s, From: %s, To: %s, Pos: %d",
		userID, req.QuadroID, req.CardID, req.SourceColumnID, req.TargetColumnID, req.Posicao)

	if err := h.kanbanService.MoveCard(req.QuadroID, req.CardID, req.SourceColumnID, req.TargetColumnID, req.Posicao, userID); err != nil {
		log.Printf("[KANBAN] MoveCard - Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao mover card"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Card movido com sucesso"})
}

// GetMetadata retorna metadados do quadro
func (h *KanbanHandler) GetMetadata(c *gin.Context) {
	userID := c.GetString("user_id")
	quadroID := c.Param("id")

	log.Printf("[KANBAN] GetMetadata - UserID: %s, QuadroID: %s", userID, quadroID)

	metadata, err := h.kanbanService.GetMetadata(quadroID, userID)
	if err != nil {
		log.Printf("[KANBAN] GetMetadata - Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar metadados"})
		return
	}

	c.JSON(http.StatusOK, metadata)
}

// UpdateColumnColor atualiza a cor de uma coluna
func (h *KanbanHandler) UpdateColumnColor(c *gin.Context) {
	userID := c.GetString("user_id")
	colunaID := c.Param("colunaId")

	var req struct {
		Cor string `json:"cor" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[KANBAN] UpdateColumnColor - UserID: %s, ColunaID: %s, NewColor: %s", userID, colunaID, req.Cor)

	if err := h.kanbanService.UpdateColumnColor(colunaID, userID, req.Cor); err != nil {
		log.Printf("[KANBAN] UpdateColumnColor - Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar cor da coluna"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cor da coluna atualizada com sucesso"})
}

// ReorderColumns reordena as colunas de um quadro
func (h *KanbanHandler) ReorderColumns(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		QuadroID    string                   `json:"quadroId" binding:"required"`
		ColumnOrder []map[string]interface{} `json:"columnOrder" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[KANBAN] ReorderColumns - UserID: %s, QuadroID: %s, Columns: %d", userID, req.QuadroID, len(req.ColumnOrder))

	if err := h.kanbanService.ReorderColumns(req.QuadroID, userID, req.ColumnOrder); err != nil {
		log.Printf("[KANBAN] ReorderColumns - Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao reordenar colunas"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Colunas reordenadas com sucesso"})
}

// ProxyToWAHA faz proxy das requisições do frontend para o WAHA interno
func (h *WhatsAppHandler) ProxyToWAHA(c *gin.Context) {
	// Construir URL do WAHA
	wahaURL := h.whatsappService.GetWAHAURL()
	targetURL := wahaURL + c.Request.URL.Path

	// Preservar query parameters
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	log.Printf("[PROXY] Proxying %s %s to %s", c.Request.Method, c.Request.URL.Path, targetURL)

	// Fazer requisição para WAHA
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		log.Printf("[PROXY] Erro ao criar requisição: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno"})
		return
	}

	// Adicionar headers necessários
	req.Header.Set("X-Api-Key", os.Getenv("WHATSAPP_API_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	// Fazer requisição
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[PROXY] Erro na requisição para WAHA: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Erro ao conectar com WAHA"})
		return
	}
	defer resp.Body.Close()

	// Ler resposta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[PROXY] Erro ao ler resposta: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno"})
		return
	}

	// Copiar headers de resposta
	for k, v := range resp.Header {
		if len(v) > 0 {
			c.Header(k, v[0])
		}
	}

	// Retornar resposta
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}
