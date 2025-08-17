package handlers

import (
	"log"
	"net/http"
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
	// TODO: Implementar listagem de sessões
	c.JSON(http.StatusOK, gin.H{"message": "Lista de sessões WhatsApp"})
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
	// TODO: Implementar webhook do WhatsApp
	c.JSON(http.StatusOK, gin.H{"message": "Webhook WhatsApp processado"})
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
