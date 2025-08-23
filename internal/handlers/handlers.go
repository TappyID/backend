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

	"tappyone/internal/models"
	"tappyone/internal/services"

	"github.com/gin-gonic/gin"
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
	// TODO: Implementar listagem de usuários
	c.JSON(http.StatusOK, gin.H{"message": "Lista de usuários"})
}

func (h *UserHandler) Create(c *gin.Context) {
	// TODO: Implementar criação de usuário
	c.JSON(http.StatusOK, gin.H{"message": "Criação de usuário"})
}

func (h *UserHandler) GetByID(c *gin.Context) {
	// TODO: Implementar busca por ID
	c.JSON(http.StatusOK, gin.H{"message": "Busca por ID"})
}

func (h *UserHandler) Update(c *gin.Context) {
	// TODO: Implementar atualização
	c.JSON(http.StatusOK, gin.H{"message": "Atualização"})
}

func (h *UserHandler) Delete(c *gin.Context) {
	// TODO: Implementar exclusão
	c.JSON(http.StatusOK, gin.H{"message": "Exclusão"})
}

// WhatsAppHandler gerencia WhatsApp
type WhatsAppHandler struct {
	whatsappService *services.WhatsAppService
}

func NewWhatsAppHandler(whatsappService *services.WhatsAppService) *WhatsAppHandler {
	return &WhatsAppHandler{whatsappService: whatsappService}
}

func (h *WhatsAppHandler) CreateSession(c *gin.Context) {
	// TODO: Implementar criação de sessão
	c.JSON(http.StatusOK, gin.H{"message": "Criação de sessão WhatsApp"})
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
	// TODO: Implementar obtenção de QR Code
	c.JSON(http.StatusOK, gin.H{"message": "QR Code WhatsApp"})
}

func (h *WhatsAppHandler) WebhookHandler(c *gin.Context) {
	var webhookData map[string]interface{}
	if err := c.ShouldBindJSON(&webhookData); err != nil {
		log.Printf("Erro ao fazer bind do webhook: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	log.Printf("Webhook recebido: %+v", webhookData)

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

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processado"})
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
	publicURL := fmt.Sprintf("https://server.tappy.id/api/files/%s", finalFilename)
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
	return fmt.Sprintf("https://server.tappy.id/api/files/%s", uniqueFilename), nil
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
