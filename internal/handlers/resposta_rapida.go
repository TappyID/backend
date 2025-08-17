package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"tappyone/internal/services"
)

type RespostaRapidaHandler struct {
	service *services.RespostaRapidaService
}

func NewRespostaRapidaHandler(service *services.RespostaRapidaService) *RespostaRapidaHandler {
	return &RespostaRapidaHandler{service: service}
}

// ===== CATEGORIAS =====

// CreateCategoria cria uma nova categoria de resposta rápida
func (h *RespostaRapidaHandler) CreateCategoria(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	usuarioID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}

	var req struct {
		Nome      string `json:"nome" binding:"required"`
		Descricao string `json:"descricao"`
		Cor       string `json:"cor" binding:"required"`
		Icone     string `json:"icone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	categoria, err := h.service.CreateCategoria(usuarioID, req.Nome, req.Descricao, req.Cor, req.Icone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, categoria)
}

// GetCategorias lista todas as categorias do usuário
func (h *RespostaRapidaHandler) GetCategorias(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	usuarioID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}

	categorias, err := h.service.GetCategoriasByUsuario(usuarioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categorias)
}

// UpdateCategoria atualiza uma categoria
func (h *RespostaRapidaHandler) UpdateCategoria(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var req struct {
		Nome      string `json:"nome" binding:"required"`
		Descricao string `json:"descricao"`
		Cor       string `json:"cor" binding:"required"`
		Icone     string `json:"icone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.UpdateCategoria(id, req.Nome, req.Descricao, req.Cor, req.Icone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "categoria atualizada com sucesso"})
}

// DeleteCategoria deleta uma categoria
func (h *RespostaRapidaHandler) DeleteCategoria(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	err = h.service.DeleteCategoria(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "categoria deletada com sucesso"})
}

// ===== RESPOSTAS RÁPIDAS =====

// CreateRespostaRapida cria uma nova resposta rápida
func (h *RespostaRapidaHandler) CreateRespostaRapida(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	usuarioID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}

	var req services.CreateRespostaRapidaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Definir o usuário ID na request
	req.UsuarioID = usuarioID

	resposta, err := h.service.CreateRespostaRapida(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resposta)
}

// GetRespostasRapidas lista todas as respostas rápidas do usuário
func (h *RespostaRapidaHandler) GetRespostasRapidas(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	usuarioID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}

	respostas, err := h.service.GetRespostasRapidasByUsuario(usuarioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, respostas)
}

// GetRespostaRapida busca uma resposta rápida específica
func (h *RespostaRapidaHandler) GetRespostaRapida(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	resposta, err := h.service.GetRespostaRapidaByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "resposta rápida não encontrada"})
		return
	}

	c.JSON(http.StatusOK, resposta)
}

// TogglePausarRespostaRapida pausa/despausa uma resposta rápida
func (h *RespostaRapidaHandler) TogglePausarRespostaRapida(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var req struct {
		Pausado bool `json:"pausado"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.TogglePausarRespostaRapida(id, req.Pausado)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	status := "despausada"
	if req.Pausado {
		status = "pausada"
	}

	c.JSON(http.StatusOK, gin.H{"message": "resposta rápida " + status + " com sucesso"})
}

// ExecutarRespostaRapida executa uma resposta rápida manualmente
func (h *RespostaRapidaHandler) ExecutarRespostaRapida(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	usuarioID, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var req struct {
		ChatID string `json:"chat_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.ExecutarRespostaRapida(id, req.ChatID, usuarioID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "resposta rápida executada com sucesso"})
}

// ===== AÇÕES =====

// CreateAcao cria uma nova ação para uma resposta rápida
func (h *RespostaRapidaHandler) CreateAcao(c *gin.Context) {
	respostaIDParam := c.Param("id")
	_, err := uuid.Parse(respostaIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da resposta rápida inválido"})
		return
	}

	var req services.CreateAcaoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implementar criação de ação individual
	c.JSON(http.StatusNotImplemented, gin.H{"error": "funcionalidade não implementada ainda"})
}

// GetAcoes lista todas as ações de uma resposta rápida
func (h *RespostaRapidaHandler) GetAcoes(c *gin.Context) {
	respostaIDParam := c.Param("id")
	respostaID, err := uuid.Parse(respostaIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da resposta rápida inválido"})
		return
	}

	resposta, err := h.service.GetRespostaRapidaByID(respostaID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "resposta rápida não encontrada"})
		return
	}

	c.JSON(http.StatusOK, resposta.Acoes)
}

// ReorderAcoes reordena as ações de uma resposta rápida
func (h *RespostaRapidaHandler) ReorderAcoes(c *gin.Context) {
	respostaIDParam := c.Param("id")
	_, err := uuid.Parse(respostaIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da resposta rápida inválido"})
		return
	}

	var req struct {
		AcoesOrdem []uuid.UUID `json:"acoes_ordem" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implementar reordenação de ações
	c.JSON(http.StatusNotImplemented, gin.H{"error": "funcionalidade não implementada ainda"})
}

// ===== EXECUÇÕES =====

// GetExecucoes lista execuções de respostas rápidas
func (h *RespostaRapidaHandler) GetExecucoes(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	_, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}

	// Parâmetros de filtro opcionais (não utilizados ainda)
	_ = c.Query("resposta_id")
	_ = c.Query("chat_id")
	_ = c.Query("status")
	limitParam := c.DefaultQuery("limit", "50")
	offsetParam := c.DefaultQuery("offset", "0")

	_, err = strconv.Atoi(limitParam)
	if err != nil {
		// limit = 50
	}

	_, err = strconv.Atoi(offsetParam)
	if err != nil {
		// offset = 0
	}

	// TODO: Implementar busca de execuções com filtros
	c.JSON(http.StatusNotImplemented, gin.H{"error": "funcionalidade não implementada ainda"})
}

// GetEstatisticas retorna estatísticas das respostas rápidas
func (h *RespostaRapidaHandler) GetEstatisticas(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	_, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}

	// TODO: Implementar estatísticas
	c.JSON(http.StatusNotImplemented, gin.H{"error": "funcionalidade não implementada ainda"})
}

// ===== AGENDAMENTOS =====

// GetAgendamentos lista agendamentos ativos
func (h *RespostaRapidaHandler) GetAgendamentos(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	_, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}

	// TODO: Implementar busca de agendamentos
	c.JSON(http.StatusNotImplemented, gin.H{"error": "funcionalidade não implementada ainda"})
}

// PausarAgendamento pausa/despausa um agendamento
func (h *RespostaRapidaHandler) PausarAgendamento(c *gin.Context) {
	idParam := c.Param("id")
	_, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var req struct {
		Pausado bool `json:"pausado"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implementar pausar/despausar agendamento
	c.JSON(http.StatusNotImplemented, gin.H{"error": "funcionalidade não implementada ainda"})
}

// ===== WEBHOOK PARA MENSAGENS =====

// ProcessarMensagemWebhook processa mensagens recebidas via webhook
func (h *RespostaRapidaHandler) ProcessarMensagemWebhook(c *gin.Context) {
	var req struct {
		Event   string `json:"event"`
		Session string `json:"session"`
		Payload struct {
			ID   string `json:"id"`
			From string `json:"from"`
			Body string `json:"body"`
			Type string `json:"type"`
		} `json:"payload"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar se é uma mensagem recebida
	if req.Event != "message" || req.Payload.Type != "text" {
		c.JSON(http.StatusOK, gin.H{"message": "evento ignorado"})
		return
	}

	// Extrair user_id da session (assumindo formato "user_{id}")
	sessionParts := strings.Split(req.Session, "_")
	if len(sessionParts) != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "formato de sessão inválido"})
		return
	}

	usuarioID, err := uuid.Parse(sessionParts[1])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido na sessão"})
		return
	}

	// Processar mensagem
	err = h.service.ProcessarMensagemRecebida(
		usuarioID,
		req.Payload.From,
		req.Payload.Body,
		"", // TODO: Extrair nome do contato
		req.Payload.From,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "mensagem processada com sucesso"})
}

// ===== COMANDOS SLASH =====

// ProcessarComandoSlash processa comandos slash no chat
func (h *RespostaRapidaHandler) ProcessarComandoSlash(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	_, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}

	var req struct {
		ChatID   string `json:"chat_id" binding:"required"`
		Comando  string `json:"comando" binding:"required"`
		Mensagem string `json:"mensagem"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implementar processamento de comandos slash
	// Exemplos: /resposta1, /bom-dia, /pix, etc.
	
	c.JSON(http.StatusNotImplemented, gin.H{"error": "funcionalidade não implementada ainda"})
}
