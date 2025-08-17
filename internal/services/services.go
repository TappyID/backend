package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"tappyone/internal/config"
	"tappyone/internal/models"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// UserService gerencia operações de usuários
type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetByID(id string) (*models.Usuario, error) {
	var usuario models.Usuario
	if err := s.db.First(&usuario, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &usuario, nil
}

func (s *UserService) Create(usuario *models.Usuario) error {
	return s.db.Create(usuario).Error
}

func (s *UserService) Update(usuario *models.Usuario) error {
	return s.db.Save(usuario).Error
}

// WhatsAppService gerencia integração com WhatsApp
type WhatsAppService struct {
	db     *gorm.DB
	config *config.Config
	client *http.Client
}

func NewWhatsAppService(db *gorm.DB, config *config.Config) *WhatsAppService {
	return &WhatsAppService{
		db:     db,
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *WhatsAppService) CreateSession(session *models.SessaoWhatsApp) error {
	return s.db.Create(session).Error
}

func (s *WhatsAppService) GetSessionByID(id string) (*models.SessaoWhatsApp, error) {
	var session models.SessaoWhatsApp
	if err := s.db.First(&session, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// WAHA API Methods
func (s *WhatsAppService) makeWAHARequest(method, endpoint, sessionName string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	wahaURL := os.Getenv("WAHA_API_URL")
	wahaKey := os.Getenv("WAHA_API_KEY")
	
	url := fmt.Sprintf("%s%s", wahaURL, endpoint)
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", wahaKey)
	if sessionName != "" {
		q := req.URL.Query()
		q.Add("session", sessionName)
		req.URL.RawQuery = q.Encode()
	}

	return s.client.Do(req)
}

func (s *WhatsAppService) GetChats(sessionName string) (interface{}, error) {
	log.Printf("[WHATSAPP] GET /chats - Starting request")
	log.Printf("[WHATSAPP] GET /chats - UserID: %s, SessionName: %s", "", sessionName)
	
	// Usar chats completo para ter todas as conversas (não apenas overview)
	endpoint := fmt.Sprintf("/%s/chats", sessionName)
	log.Printf("[WHATSAPP] GET /chats - Using endpoint: %s", endpoint)
	resp, err := s.makeWAHARequest("GET", endpoint, "", nil) // sessionName já está no endpoint
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *WhatsAppService) GetContacts(sessionName string) (interface{}, error) {
	// Endpoint correto: /contacts/all?session={sessionName}
	endpoint := fmt.Sprintf("/contacts/all?session=%s", sessionName)
	resp, err := s.makeWAHARequest("GET", endpoint, "", nil) // sessionName já está no endpoint
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *WhatsAppService) GetChatMessages(sessionName, chatID string) (interface{}, error) {
	// Endpoint correto: /messages?chatId={chatId}&session={session} (sem /api/ pois já está na base URL)
	// Buscar mais mensagens e ordenar por mais recentes
	endpoint := fmt.Sprintf("/messages?chatId=%s&session=%s&limit=50&downloadMedia=false", chatID, sessionName)
	log.Printf("[WHATSAPP] GetChatMessages - sessionName: %s, chatID: %s", sessionName, chatID)
	log.Printf("[WHATSAPP] GetChatMessages - endpoint: %s", endpoint)
	
	resp, err := s.makeWAHARequest("GET", endpoint, "", nil) // sessionName já está no endpoint
	if err != nil {
		log.Printf("[WHATSAPP] GetChatMessages - makeWAHARequest error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[WHATSAPP] GetChatMessages - WAHA API response status: %d", resp.StatusCode)
	
	if resp.StatusCode != http.StatusOK {
		// Ler o corpo da resposta para debug
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[WHATSAPP] GetChatMessages - WAHA API error response: %s", string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[WHATSAPP] GetChatMessages - JSON decode error: %v", err)
		return nil, err
	}
	
	log.Printf("[WHATSAPP] GetChatMessages - Success, returning %d messages", len(result.([]interface{})))
	return result, nil
}

func (s *WhatsAppService) SendMessage(sessionName, chatID, text string) (interface{}, error) {
	// Endpoint correto da WAHA API: /sendText (sem /api/ pois já está na base URL)
	endpoint := "/sendText"
	body := map[string]string{
		"session": sessionName,
		"chatId":  chatID,
		"text":    text,
	}
	
	log.Printf("[WHATSAPP] SendMessage - sessionName: %s, chatID: %s, text: %s", sessionName, chatID, text)
	log.Printf("[WHATSAPP] SendMessage - endpoint: %s", endpoint)
	
	resp, err := s.makeWAHARequest("POST", endpoint, "", body) // sessionName já está no body
	if err != nil {
		log.Printf("[WHATSAPP] SendMessage - makeWAHARequest error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[WHATSAPP] SendMessage - WAHA API response status: %d", resp.StatusCode)
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// Ler o corpo da resposta para debug
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[WHATSAPP] SendMessage - WAHA API error response: %s", string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[WHATSAPP] SendMessage - JSON decode error: %v", err)
		return nil, err
	}
	
	log.Printf("[WHATSAPP] SendMessage - Success!")
	return result, nil
}

// GetPresence obtém informações de presença para todos os chats
func (s *WhatsAppService) GetPresence(sessionName string) (interface{}, error) {
	endpoint := fmt.Sprintf("/%s/presence", sessionName)
	log.Printf("[WHATSAPP] GetPresence - sessionName: %s, endpoint: %s", sessionName, endpoint)
	
	resp, err := s.makeWAHARequest("GET", endpoint, "", nil)
	if err != nil {
		log.Printf("[WHATSAPP] GetPresence - makeWAHARequest error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[WHATSAPP] GetPresence - WAHA API response status: %d", resp.StatusCode)
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[WHATSAPP] GetPresence - WAHA API error response: %s", string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[WHATSAPP] GetPresence - JSON decode error: %v", err)
		return nil, err
	}
	
	log.Printf("[WHATSAPP] GetPresence - Success!")
	return result, nil
}

// MarkAsRead marca mensagens como lidas
func (s *WhatsAppService) MarkAsRead(sessionName, chatID string) (interface{}, error) {
	endpoint := fmt.Sprintf("/%s/chats/%s/messages/read", sessionName, chatID)
	log.Printf("[WHATSAPP] MarkAsRead - sessionName: %s, chatID: %s, endpoint: %s", sessionName, chatID, endpoint)
	
	resp, err := s.makeWAHARequest("POST", endpoint, "", nil)
	if err != nil {
		log.Printf("[WHATSAPP] MarkAsRead - makeWAHARequest error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[WHATSAPP] MarkAsRead - WAHA API response status: %d", resp.StatusCode)
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[WHATSAPP] MarkAsRead - WAHA API error response: %s", string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[WHATSAPP] MarkAsRead - JSON decode error: %v", err)
		return nil, err
	}
	
	log.Printf("[WHATSAPP] MarkAsRead - Success!")
	return result, nil
}

func (s *WhatsAppService) SetTyping(sessionName, chatID, presence string) (interface{}, error) {
	endpoint := fmt.Sprintf("/%s/presence", sessionName)
	body := map[string]string{
		"chatId":   chatID,
		"presence": presence, // "typing", "paused", etc.
	}
	
	log.Printf("[WHATSAPP] SetTyping - sessionName: %s, chatID: %s, presence: %s", sessionName, chatID, presence)
	log.Printf("[WHATSAPP] SetTyping - endpoint: %s", endpoint)
	
	resp, err := s.makeWAHARequest("POST", endpoint, "", body)
	if err != nil {
		log.Printf("[WHATSAPP] SetTyping - makeWAHARequest error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[WHATSAPP] SetTyping - WAHA API response status: %d", resp.StatusCode)
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[WHATSAPP] SetTyping - WAHA API error response: %s", string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[WHATSAPP] SetTyping - JSON decode error: %v", err)
		return nil, err
	}
	
	log.Printf("[WHATSAPP] SetTyping - Success!")
	return result, nil
}

// SendImage envia uma imagem
func (s *WhatsAppService) SendImage(sessionName, chatID, imageURL, caption string) (interface{}, error) {
	endpoint := "/sendImage"
	body := map[string]interface{}{
		"session": sessionName,
		"chatId":  chatID,
		"file": map[string]string{
			"url": imageURL,
		},
		"caption": caption,
	}
	
	log.Printf("[WHATSAPP] SendImage - sessionName: %s, chatID: %s, imageURL: %s", sessionName, chatID, imageURL)
	
	resp, err := s.makeWAHARequest("POST", endpoint, "", body)
	if err != nil {
		log.Printf("[WHATSAPP] SendImage - makeWAHARequest error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[WHATSAPP] SendImage - WAHA API response status: %d", resp.StatusCode)
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[WHATSAPP] SendImage - WAHA API error response: %s", string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[WHATSAPP] SendImage - JSON decode error: %v", err)
		return nil, err
	}
	
	log.Printf("[WHATSAPP] SendImage - Success!")
	return result, nil
}

// SendFile envia um arquivo
func (s *WhatsAppService) SendFile(sessionName, chatID, fileURL, filename, caption string) (interface{}, error) {
	endpoint := "/sendFile"
	body := map[string]interface{}{
		"session": sessionName,
		"chatId":  chatID,
		"file": map[string]string{
			"url": fileURL,
		},
		"filename": filename,
		"caption":  caption,
	}
	
	log.Printf("[WHATSAPP] SendFile - sessionName: %s, chatID: %s, fileURL: %s, filename: %s", sessionName, chatID, fileURL, filename)
	
	resp, err := s.makeWAHARequest("POST", endpoint, "", body)
	if err != nil {
		log.Printf("[WHATSAPP] SendFile - makeWAHARequest error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[WHATSAPP] SendFile - WAHA API response status: %d", resp.StatusCode)
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[WHATSAPP] SendFile - WAHA API error response: %s", string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[WHATSAPP] SendFile - JSON decode error: %v", err)
		return nil, err
	}
	
	log.Printf("[WHATSAPP] SendFile - Success!")
	return result, nil
}

// SendVoice envia um áudio
func (s *WhatsAppService) SendVoice(sessionName, chatID, audioURL string) (interface{}, error) {
	endpoint := "/sendVoice"
	body := map[string]interface{}{
		"session": sessionName,
		"chatId":  chatID,
		"file": map[string]string{
			"url": audioURL,
		},
	}
	
	log.Printf("[WHATSAPP] SendVoice - sessionName: %s, chatID: %s, audioURL: %s", sessionName, chatID, audioURL)
	
	resp, err := s.makeWAHARequest("POST", endpoint, "", body)
	if err != nil {
		log.Printf("[WHATSAPP] SendVoice - makeWAHARequest error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[WHATSAPP] SendVoice - WAHA API response status: %d", resp.StatusCode)
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[WHATSAPP] SendVoice - WAHA API error response: %s", string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[WHATSAPP] SendVoice - JSON decode error: %v", err)
		return nil, err
	}
	
	log.Printf("[WHATSAPP] SendVoice - Success!")
	return result, nil
}

// SendVideo envia um vídeo
func (s *WhatsAppService) SendVideo(sessionName, chatID, videoURL, caption string) (interface{}, error) {
	endpoint := "/sendVideo"
	body := map[string]interface{}{
		"session": sessionName,
		"chatId":  chatID,
		"file": map[string]string{
			"url": videoURL,
		},
		"caption": caption,
	}
	
	log.Printf("[WHATSAPP] SendVideo - sessionName: %s, chatID: %s, videoURL: %s", sessionName, chatID, videoURL)
	
	resp, err := s.makeWAHARequest("POST", endpoint, "", body)
	if err != nil {
		log.Printf("[WHATSAPP] SendVideo - makeWAHARequest error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[WHATSAPP] SendVideo - WAHA API response status: %d", resp.StatusCode)
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[WHATSAPP] SendVideo - WAHA API error response: %s", string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[WHATSAPP] SendVideo - JSON decode error: %v", err)
		return nil, err
	}
	
	log.Printf("[WHATSAPP] SendVideo - Success!")
	return result, nil
}

// SendVoiceFile envia áudio diretamente como arquivo
func (s *WhatsAppService) SendVoiceFile(sessionName, chatID, filePath string) (interface{}, error) {
	// Abrir o arquivo
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[WHATSAPP] SendVoiceFile - Error opening file: %v", err)
		return nil, err
	}
	defer file.Close()

	// Criar buffer para multipart
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Adicionar campos do formulário
	writer.WriteField("session", sessionName)
	writer.WriteField("chatId", chatID)

	// Adicionar arquivo
	filename := filepath.Base(filePath)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		log.Printf("[WHATSAPP] SendVoiceFile - Error creating form file: %v", err)
		return nil, err
	}

	// Copiar arquivo para o form
	if _, err := io.Copy(part, file); err != nil {
		log.Printf("[WHATSAPP] SendVoiceFile - Error copying file: %v", err)
		return nil, err
	}

	// Fechar writer
	writer.Close()

	// Fazer requisição - usar as mesmas variáveis do sistema antigo
	wahaURL := os.Getenv("WAHA_API_URL")
	if wahaURL == "" {
		wahaURL = s.config.WhatsAppAPIURL + "/api"
	}
	url := wahaURL + "/sendVoice"
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		log.Printf("[WHATSAPP] SendVoiceFile - Error creating request: %v", err)
		return nil, err
	}

	// Headers
	wahaKey := os.Getenv("WAHA_API_KEY")
	if wahaKey == "" {
		wahaKey = s.config.WhatsAppAPIToken
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-Api-Key", wahaKey)

	log.Printf("[WHATSAPP] SendVoiceFile - sessionName: %s, chatID: %s, filePath: %s", sessionName, chatID, filePath)

	// Executar requisição
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[WHATSAPP] SendVoiceFile - HTTP request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("[WHATSAPP] SendVoiceFile - WAHA API response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("[WHATSAPP] SendVoiceFile - WAHA API error response: %s", string(body))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[WHATSAPP] SendVoiceFile - JSON decode error: %v", err)
		return nil, err
	}

	log.Printf("[WHATSAPP] SendVoiceFile - Success!")
	return result, nil
}

// KanbanService gerencia quadros Kanban
type KanbanService struct {
	db *gorm.DB
}

func NewKanbanService(db *gorm.DB) *KanbanService {
	return &KanbanService{db: db}
}

func (s *KanbanService) CreateQuadro(quadro *models.Quadro) error {
	log.Printf("[KANBAN] CreateQuadro - Before create: ID=%s, Nome=%s, UsuarioID=%s, Ativo=%v", quadro.ID, quadro.Nome, quadro.UsuarioID, quadro.Ativo)
	err := s.db.Create(quadro).Error
	if err != nil {
		log.Printf("[KANBAN] CreateQuadro - Error: %v", err)
		return err
	}
	
	// Criar colunas padrão
	corNovo := "#ef4444"
	corAndamento := "#f59e0b"
	corAguardando := "#3b82f6"
	corConcluido := "#10b981"
	
	colunasDefault := []models.Coluna{
		{Nome: "Novo", Posicao: 0, QuadroID: quadro.ID, Cor: &corNovo, Ativo: true},
		{Nome: "Em Andamento", Posicao: 1, QuadroID: quadro.ID, Cor: &corAndamento, Ativo: true},
		{Nome: "Aguardando", Posicao: 2, QuadroID: quadro.ID, Cor: &corAguardando, Ativo: true},
		{Nome: "Concluído", Posicao: 3, QuadroID: quadro.ID, Cor: &corConcluido, Ativo: true},
	}
	
	for _, coluna := range colunasDefault {
		if err := s.db.Create(&coluna).Error; err != nil {
			log.Printf("[KANBAN] CreateQuadro - Error creating default column %s: %v", coluna.Nome, err)
			return err
		}
		log.Printf("[KANBAN] CreateQuadro - Created default column: ID=%s, Nome=%s", coluna.ID, coluna.Nome)
	}
	
	log.Printf("[KANBAN] CreateQuadro - After create: ID=%s, Nome=%s, UsuarioID=%s, Ativo=%v", quadro.ID, quadro.Nome, quadro.UsuarioID, quadro.Ativo)
	return nil
}

func (s *KanbanService) GetQuadrosByUser(userID string) ([]models.Quadro, error) {
	var quadros []models.Quadro
	log.Printf("[KANBAN] GetQuadrosByUser - Searching for userID: %s", userID)
	
	// Primeiro, vamos tentar buscar todos os quadros para debug
	var allQuadros []models.Quadro
	s.db.Find(&allQuadros)
	log.Printf("[KANBAN] GetQuadrosByUser - Total quadros in DB: %d", len(allQuadros))
	for i, q := range allQuadros {
		log.Printf("[KANBAN] GetQuadrosByUser - Quadro %d: ID=%s, Nome=%s, UsuarioID=%s, Ativo=%v", i, q.ID, q.Nome, q.UsuarioID, q.Ativo)
	}
	
	// Agora buscar com a query original
	if err := s.db.Where("usuario_id = ? AND ativo = ?", userID, true).Order("posicao ASC").Find(&quadros).Error; err != nil {
		log.Printf("[KANBAN] GetQuadrosByUser - Error: %v", err)
		return nil, err
	}
	log.Printf("[KANBAN] GetQuadrosByUser - Found %d quadros for user %s", len(quadros), userID)
	return quadros, nil
}

func (s *KanbanService) GetQuadroByID(quadroID, userID string) (*models.Quadro, error) {
	var quadro models.Quadro
	if err := s.db.Preload("Colunas", "ativo = ?", true).Where("id = ? AND usuario_id = ? AND ativo = ?", quadroID, userID, true).First(&quadro).Error; err != nil {
		return nil, err
	}
	
	log.Printf("[KANBAN] GetQuadroByID - Quadro: %s, Colunas encontradas: %d", quadro.Nome, len(quadro.Colunas))
	for i, col := range quadro.Colunas {
		log.Printf("[KANBAN] GetQuadroByID - Coluna %d: ID=%s, Nome=%s", i, col.ID, col.Nome)
	}
	
	// Se o quadro não tem colunas, criar colunas padrão
	if len(quadro.Colunas) == 0 {
		log.Printf("[KANBAN] GetQuadroByID - Quadro sem colunas, criando colunas padrão")
		if err := s.createDefaultColumns(quadro.ID); err != nil {
			log.Printf("[KANBAN] GetQuadroByID - Error creating default columns: %v", err)
		} else {
			// Recarregar o quadro com as colunas criadas
			if err := s.db.Preload("Colunas", "ativo = ?", true).Where("id = ?", quadro.ID).First(&quadro).Error; err != nil {
				log.Printf("[KANBAN] GetQuadroByID - Error reloading quadro: %v", err)
			} else {
				log.Printf("[KANBAN] GetQuadroByID - Colunas padrão criadas: %d", len(quadro.Colunas))
				for i, col := range quadro.Colunas {
					log.Printf("[KANBAN] GetQuadroByID - Nova Coluna %d: ID=%s, Nome=%s", i, col.ID, col.Nome)
				}
			}
		}
	}
	
	return &quadro, nil
}

func (s *KanbanService) createDefaultColumns(quadroID string) error {
	// Criar colunas padrão
	corNovo := "#ef4444"
	corAndamento := "#f59e0b"
	corAguardando := "#3b82f6"
	corConcluido := "#10b981"
	
	colunasDefault := []models.Coluna{
		{Nome: "Novo", Posicao: 0, QuadroID: quadroID, Cor: &corNovo, Ativo: true},
		{Nome: "Em Andamento", Posicao: 1, QuadroID: quadroID, Cor: &corAndamento, Ativo: true},
		{Nome: "Aguardando", Posicao: 2, QuadroID: quadroID, Cor: &corAguardando, Ativo: true},
		{Nome: "Concluído", Posicao: 3, QuadroID: quadroID, Cor: &corConcluido, Ativo: true},
	}
	
	for _, coluna := range colunasDefault {
		if err := s.db.Create(&coluna).Error; err != nil {
			log.Printf("[KANBAN] createDefaultColumns - Error creating column %s: %v", coluna.Nome, err)
			return err
		}
		log.Printf("[KANBAN] createDefaultColumns - Created column: ID=%s, Nome=%s", coluna.ID, coluna.Nome)
	}
	
	return nil
}

func (s *KanbanService) UpdateQuadro(quadroID, userID string, nome, cor, descricao *string, posicao *int, ativo *bool) (*models.Quadro, error) {
	var quadro models.Quadro
	if err := s.db.Where("id = ? AND usuario_id = ?", quadroID, userID).First(&quadro).Error; err != nil {
		return nil, err
	}
	
	// Atualizar apenas campos fornecidos
	if nome != nil {
		quadro.Nome = *nome
	}
	if cor != nil {
		quadro.Cor = *cor
	}
	if descricao != nil {
		quadro.Descricao = descricao
	}
	if posicao != nil {
		quadro.Posicao = *posicao
	}
	if ativo != nil {
		quadro.Ativo = *ativo
	}
	
	if err := s.db.Save(&quadro).Error; err != nil {
		return nil, err
	}
	
	return &quadro, nil
}

func (s *KanbanService) DeleteQuadro(quadroID, userID string) error {
	// Soft delete - marcar como inativo
	result := s.db.Model(&models.Quadro{}).Where("id = ? AND usuario_id = ?", quadroID, userID).Update("ativo", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("quadro não encontrado")
	}
	return nil
}

// EditColumn edita o nome de uma coluna
func (s *KanbanService) EditColumn(quadroID, colunaID, novoNome, userID string) error {
	log.Printf("[KANBAN_SERVICE] EditColumn - QuadroID: %s, ColunaID: %s, NovoNome: %s, UserID: %s", quadroID, colunaID, novoNome, userID)
	
	// Verificar se o quadro pertence ao usuário
	var quadro models.Quadro
	if err := s.db.Where("id = ? AND usuario_id = ? AND ativo = ?", quadroID, userID, true).First(&quadro).Error; err != nil {
		return fmt.Errorf("quadro não encontrado")
	}
	
	// Atualizar nome da coluna
	result := s.db.Model(&models.Coluna{}).Where("id = ? AND quadro_id = ? AND ativo = ?", colunaID, quadroID, true).Update("nome", novoNome)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("coluna não encontrada")
	}
	
	log.Printf("[KANBAN_SERVICE] EditColumn - Coluna atualizada com sucesso")
	return nil
}

// DeleteColumn exclui uma coluna
func (s *KanbanService) DeleteColumn(quadroID, colunaID, userID string) error {
	log.Printf("[KANBAN_SERVICE] DeleteColumn - QuadroID: %s, ColunaID: %s, UserID: %s", quadroID, colunaID, userID)
	
	// Verificar se o quadro pertence ao usuário
	var quadro models.Quadro
	if err := s.db.Where("id = ? AND usuario_id = ? AND ativo = ?", quadroID, userID, true).First(&quadro).Error; err != nil {
		return fmt.Errorf("quadro não encontrado")
	}
	
	// Soft delete da coluna
	result := s.db.Model(&models.Coluna{}).Where("id = ? AND quadro_id = ? AND ativo = ?", colunaID, quadroID, true).Update("ativo", false)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("coluna não encontrada")
	}
	
	// Também marcar cards da coluna como inativos
	s.db.Model(&models.Card{}).Where("coluna_id = ? AND ativo = ?", colunaID, true).Update("ativo", false)
	
	log.Printf("[KANBAN_SERVICE] DeleteColumn - Coluna excluída com sucesso")
	return nil
}

// CreateColumn cria uma nova coluna
func (s *KanbanService) CreateColumn(quadroID, nome string, cor *string, posicao int, userID string) (*models.Coluna, error) {
	log.Printf("[KANBAN_SERVICE] CreateColumn - QuadroID: %s, Nome: %s, Posicao: %d, UserID: %s", quadroID, nome, posicao, userID)
	
	// Verificar se o quadro pertence ao usuário
	var quadro models.Quadro
	if err := s.db.Where("id = ? AND usuario_id = ? AND ativo = ?", quadroID, userID, true).First(&quadro).Error; err != nil {
		return nil, fmt.Errorf("quadro não encontrado")
	}
	
	// Criar nova coluna
	coluna := models.Coluna{
		Nome:     nome,
		Cor:      cor,
		Posicao:  posicao,
		QuadroID: quadroID,
		Ativo:    true,
	}
	
	if err := s.db.Create(&coluna).Error; err != nil {
		log.Printf("[KANBAN_SERVICE] CreateColumn - Error: %v", err)
		return nil, err
	}
	
	log.Printf("[KANBAN_SERVICE] CreateColumn - Coluna criada com sucesso: ID=%s, Nome=%s", coluna.ID, coluna.Nome)
	return &coluna, nil
}

// MoveCard move um card entre colunas
func (s *KanbanService) MoveCard(quadroID, cardID, sourceColumnID, targetColumnID string, posicao int, userID string) error {
	log.Printf("[KANBAN_SERVICE] MoveCard - QuadroID: %s, CardID: %s, From: %s, To: %s, Pos: %d, UserID: %s", 
		quadroID, cardID, sourceColumnID, targetColumnID, posicao, userID)
	
	// Verificar se o quadro pertence ao usuário
	var quadro models.Quadro
	if err := s.db.Where("id = ? AND usuario_id = ? AND ativo = ?", quadroID, userID, true).First(&quadro).Error; err != nil {
		return fmt.Errorf("quadro não encontrado")
	}
	
	// Primeiro, tentar encontrar o card existente
	var existingCard models.Card
	err := s.db.Where("conversa_id = ? AND ativo = ?", cardID, true).First(&existingCard).Error
	
	if err != nil {
		// Card não existe, criar um novo
		log.Printf("[KANBAN_SERVICE] MoveCard - Card não encontrado, criando novo card para conversa: %s", cardID)
		
		newCard := models.Card{
			Nome:       fmt.Sprintf("Conversa %s", cardID),
			Posicao:    posicao,
			ColunaID:   targetColumnID,
			ConversaID: cardID,
			Ativo:      true,
		}
		
		if err := s.db.Create(&newCard).Error; err != nil {
			log.Printf("[KANBAN_SERVICE] MoveCard - Error creating card: %v", err)
			return err
		}
		
		log.Printf("[KANBAN_SERVICE] MoveCard - Card criado com sucesso: ID=%s, ConversaID=%s", newCard.ID, newCard.ConversaID)
	} else {
		// Card existe, atualizar posição e coluna
		result := s.db.Model(&existingCard).Updates(map[string]interface{}{
			"coluna_id": targetColumnID,
			"posicao":   posicao,
		})
		
		if result.Error != nil {
			return result.Error
		}
		
		log.Printf("[KANBAN_SERVICE] MoveCard - Card atualizado com sucesso: ID=%s, ConversaID=%s", existingCard.ID, existingCard.ConversaID)
	}
	
	log.Printf("[KANBAN_SERVICE] MoveCard - Card movido com sucesso")
	return nil
}

// GetMetadata retorna metadados do quadro
func (s *KanbanService) GetMetadata(quadroID, userID string) (map[string]interface{}, error) {
	log.Printf("[KANBAN_SERVICE] GetMetadata - QuadroID: %s, UserID: %s", quadroID, userID)
	
	// Verificar se o quadro pertence ao usuário
	var quadro models.Quadro
	if err := s.db.Where("id = ? AND usuario_id = ? AND ativo = ?", quadroID, userID, true).First(&quadro).Error; err != nil {
		return nil, fmt.Errorf("quadro não encontrado")
	}
	
	// Buscar todos os cards do quadro
	var cards []models.Card
	if err := s.db.Joins("JOIN colunas ON cards.coluna_id = colunas.id").
		Where("colunas.quadro_id = ? AND cards.ativo = ?", quadroID, true).
		Find(&cards).Error; err != nil {
		log.Printf("[KANBAN_SERVICE] GetMetadata - Erro ao buscar cards: %v", err)
		cards = []models.Card{} // Continuar com array vazio se houver erro
	}
	
	// Mapear cards para metadados
	cardMetadata := make(map[string]interface{})
	for _, card := range cards {
		cardMetadata[card.ConversaID] = map[string]interface{}{
			"colunaId":         card.ColunaID,
			"posicao":          card.Posicao,
			"ultimoMovimento": card.AtualizadoEm.Format(time.RFC3339),
			"cardId":           card.ID,
			"nome":             card.Nome,
		}
	}
	
	log.Printf("[KANBAN_SERVICE] GetMetadata - Encontrados %d cards", len(cards))
	
	metadata := map[string]interface{}{
		"quadroId": quadroID,
		"cards":    cardMetadata,
		"lastSync": time.Now().Format(time.RFC3339),
		"totalCards": len(cards),
	}
	
	log.Printf("[KANBAN_SERVICE] GetMetadata - Metadados retornados")
	return metadata, nil
}

// UpdateColumnColor atualiza a cor de uma coluna
func (s *KanbanService) UpdateColumnColor(colunaID, userID, newColor string) error {
	log.Printf("[KANBAN_SERVICE] UpdateColumnColor - ColunaID: %s, UserID: %s, NewColor: %s", colunaID, userID, newColor)
	
	// Verificar se a coluna pertence ao usuário
	var coluna models.Coluna
	if err := s.db.Joins("JOIN quadros ON colunas.quadro_id = quadros.id").
		Where("colunas.id = ? AND quadros.usuario_id = ? AND colunas.ativo = ?", colunaID, userID, true).
		First(&coluna).Error; err != nil {
		return fmt.Errorf("coluna não encontrada")
	}
	
	// Atualizar a cor
	result := s.db.Model(&coluna).Update("cor", newColor)
	if result.Error != nil {
		return result.Error
	}
	
	log.Printf("[KANBAN_SERVICE] UpdateColumnColor - Cor atualizada com sucesso")
	return nil
}

// ReorderColumns reordena as colunas de um quadro
func (s *KanbanService) ReorderColumns(quadroID, userID string, columnOrder []map[string]interface{}) error {
	log.Printf("[KANBAN_SERVICE] ReorderColumns - QuadroID: %s, UserID: %s", quadroID, userID)
	
	// Verificar se o quadro pertence ao usuário
	var quadro models.Quadro
	if err := s.db.Where("id = ? AND usuario_id = ? AND ativo = ?", quadroID, userID, true).First(&quadro).Error; err != nil {
		return fmt.Errorf("quadro não encontrado")
	}
	
	// Iniciar transação
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Atualizar a posição de cada coluna
	for _, orderItem := range columnOrder {
		colunaID, ok1 := orderItem["id"].(string)
		ordem, ok2 := orderItem["ordem"].(float64) // JSON numbers são float64
		
		if !ok1 || !ok2 {
			tx.Rollback()
			return fmt.Errorf("formato inválido de ordem das colunas")
		}
		
		// Verificar se a coluna pertence ao quadro
		var coluna models.Coluna
		if err := tx.Where("id = ? AND quadro_id = ? AND ativo = ?", colunaID, quadroID, true).First(&coluna).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("coluna %s não encontrada", colunaID)
		}
		
		// Atualizar posição
		if err := tx.Model(&coluna).Update("posicao", int(ordem)).Error; err != nil {
			tx.Rollback()
			return err
		}
		
		log.Printf("[KANBAN_SERVICE] ReorderColumns - Coluna %s reordenada para posição %d", colunaID, int(ordem))
	}
	
	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		return err
	}
	
	log.Printf("[KANBAN_SERVICE] ReorderColumns - Colunas reordenadas com sucesso")
	return nil
}

// MessageService gerencia mensagens
type MessageService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewMessageService(db *gorm.DB, redis *redis.Client) *MessageService {
	return &MessageService{
		db:    db,
		redis: redis,
	}
}

func (s *MessageService) CreateMessage(message *models.Mensagem) error {
	return s.db.Create(message).Error
}

func (s *MessageService) GetMessagesByChat(chatID string) ([]models.Mensagem, error) {
	var messages []models.Mensagem
	if err := s.db.Where("conversa_id = ?", chatID).Order("timestamp ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// AIService gerencia integração com IA
type AIService struct {
	config *config.Config
}

func NewAIService(config *config.Config) *AIService {
	return &AIService{config: config}
}

func (s *AIService) GenerateResponse(prompt string, context string) (string, error) {
	// TODO: Implementar integração com DeepSeek API
	return "Resposta automática da IA", nil
}

// EmailService gerencia envio de emails
type EmailService struct {
	config *config.Config
}

func NewEmailService(config *config.Config) *EmailService {
	return &EmailService{config: config}
}

func (s *EmailService) SendEmail(to, subject, body string) error {
	// TODO: Implementar envio de email via SMTP
	return nil
}
