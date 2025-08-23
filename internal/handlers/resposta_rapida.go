package handlers

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"tappyone/internal/models"
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

	// Transformar respostas para incluir triggers extraídos do trigger_condicao
	var respostasFormatadas []map[string]interface{}
	for _, resposta := range respostas {
		respostaMap := map[string]interface{}{
			"id":                resposta.ID,
			"titulo":            resposta.Titulo,
			"descricao":         resposta.Descricao,
			"categoria_id":      resposta.CategoriaID,
			"categoria":         resposta.Categoria,
			"ativo":             resposta.Ativo,
			"trigger_tipo":      resposta.TriggerTipo,
			"agendamento_ativo": resposta.AgendamentoAtivo,
			"pausado":           resposta.Pausado,
			"ordem":             resposta.Ordem,
			"usuario_id":        resposta.UsuarioID,
			"acoes":             resposta.Acoes,
			"fallback":          resposta.Fallback,
			"created_at":        resposta.CreatedAt,
			"updated_at":        resposta.UpdatedAt,
		}

		// Extrair triggers do trigger_condicao
		var triggers []string
		if condicao, err := resposta.GetTriggerCondicao(); err == nil && condicao != nil {
			triggers = condicao.PalavrasChave
		}
		if triggers == nil {
			triggers = []string{}
		}
		respostaMap["triggers"] = triggers

		respostasFormatadas = append(respostasFormatadas, respostaMap)
	}

	c.JSON(http.StatusOK, respostasFormatadas)
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

// UpdateRespostaRapida atualiza uma resposta rápida
func (h *RespostaRapidaHandler) UpdateRespostaRapida(c *gin.Context) {
	log.Printf("[UPDATE] Iniciando update de resposta rápida")
	
	userID := c.GetString("user_id")
	if userID == "" {
		log.Printf("[UPDATE] Erro: usuário não autenticado")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}

	usuarioID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("[UPDATE] Erro: ID de usuário inválido: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
		return
	}

	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.Printf("[UPDATE] Erro: ID inválido: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	
	log.Printf("[UPDATE] UserID: %s, RespostaID: %s", usuarioID, id)

	// Verificar se a resposta existe e pertence ao usuário
	existingResposta, err := h.service.GetRespostaRapidaByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "resposta rápida não encontrada"})
		return
	}

	if existingResposta.UsuarioID != usuarioID {
		c.JSON(http.StatusForbidden, gin.H{"error": "acesso negado"})
		return
	}

	var req struct {
		Titulo             string   `json:"titulo" binding:"required"`
		Descricao          *string  `json:"descricao"`
		CategoriaID        *uuid.UUID `json:"categoria_id"`
		Triggers           []string `json:"triggers" binding:"required"`
		Ativo              bool     `json:"ativo"`
		Automatico         bool     `json:"automatico"`
		Fallback           bool     `json:"fallback"`
		AgendamentoAtivo   bool     `json:"agendamento_ativo"`
		AgendamentoConfig  interface{}  `json:"agendamento_config"`
		Acoes              []struct {
			Tipo     string      `json:"tipo" binding:"required"`
			Conteudo interface{} `json:"conteudo" binding:"required"`
			Ordem    int         `json:"ordem"`
			Ativo    bool        `json:"ativo"`
		} `json:"acoes"`
	}

	// Ler o corpo da requisição primeiro para log
	body, _ := c.GetRawData()
	log.Printf("[UPDATE] Request body: %s", string(body))
	
	// Recriar o reader para o bind
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
	
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[UPDATE] Erro no bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	log.Printf("[UPDATE] Request parsed successfully - Titulo: %s, Triggers: %v, Acoes: %d", req.Titulo, req.Triggers, len(req.Acoes))

	// Atualizar campos da resposta
	existingResposta.Titulo = req.Titulo
	existingResposta.Descricao = req.Descricao
	if req.CategoriaID != nil {
		existingResposta.CategoriaID = *req.CategoriaID
	}
	existingResposta.Ativo = req.Ativo
	existingResposta.AgendamentoAtivo = req.AgendamentoAtivo
	
	// CORRIGIR: Salvar campos de regras (automatico e fallback)
	log.Printf("[UPDATE] Setting automatico: %v, fallback: %v", req.Automatico, req.Fallback)
	if req.Automatico {
		existingResposta.TriggerTipo = "palavra_chave"
	} else {
		existingResposta.TriggerTipo = "manual"
	}
	existingResposta.Fallback = req.Fallback
	
	// Atualizar trigger condition com as palavras-chave
	log.Printf("[UPDATE] Setting triggers: %v", req.Triggers)
	triggerCondicao := &models.TriggerCondicao{
		PalavrasChave: req.Triggers,
	}
	existingResposta.SetTriggerCondicao(triggerCondicao)
	log.Printf("[UPDATE] Triggers set successfully in resposta")

	// Convert ações para []interface{}
	var acoesInterface []interface{}
	for _, acao := range req.Acoes {
		log.Printf("[UPDATE] Processing acao - Tipo: %s, Conteudo: %v, Ordem: %d", acao.Tipo, acao.Conteudo, acao.Ordem)
		acoesInterface = append(acoesInterface, map[string]interface{}{
			"tipo":     acao.Tipo,
			"conteudo": acao.Conteudo,
			"ordem":    acao.Ordem,
			"ativo":    acao.Ativo,
		})
	}
	log.Printf("[UPDATE] Total acoes to process: %d", len(acoesInterface))
	
	resposta, err := h.service.UpdateRespostaRapida(existingResposta, acoesInterface)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

func (h *RespostaRapidaHandler) DeleteRespostaRapida(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	// Verificar se o usuário tem permissão
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
		return
	}

	var userID uuid.UUID
	if userIDStr, ok := userIDInterface.(string); ok {
		userID, err = uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
			return
		}
	} else if userIDUUID, ok := userIDInterface.(uuid.UUID); ok {
		userID = userIDUUID
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Formato de ID de usuário inválido"})
		return
	}

	// Buscar resposta para verificar se pertence ao usuário
	resposta, err := h.service.GetRespostaRapidaByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Resposta rápida não encontrada"})
		return
	}

	if resposta.UsuarioID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado"})
		return
	}

	// Deletar resposta rápida
	err = h.service.DeleteRespostaRapida(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Resposta rápida excluída com sucesso"})
}

// ExecutarRespostaRapida executa uma resposta rápida manualmente
func (h *RespostaRapidaHandler) ExecutarRespostaRapida(c *gin.Context) {
	log.Printf("[ExecutarRespostaRapida] Iniciando execução de resposta rápida")
	
	userIDInterface, exists := c.Get("userID")
	if !exists {
		log.Printf("[ExecutarRespostaRapida] ERRO: userID não encontrado no contexto")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuário não autenticado"})
		return
	}
	log.Printf("[ExecutarRespostaRapida] userID encontrado: %v (tipo: %T)", userIDInterface, userIDInterface)

	var usuarioID uuid.UUID
	if userIDStr, ok := userIDInterface.(string); ok {
		var err error
		usuarioID, err = uuid.Parse(userIDStr)
		if err != nil {
			log.Printf("[ExecutarRespostaRapida] ERRO: Falha ao fazer parse do userID string: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID de usuário inválido"})
			return
		}
		log.Printf("[ExecutarRespostaRapida] userID parseado com sucesso: %s", usuarioID.String())
	} else if userIDUUID, ok := userIDInterface.(uuid.UUID); ok {
		usuarioID = userIDUUID
		log.Printf("[ExecutarRespostaRapida] userID já é UUID: %s", usuarioID.String())
	} else {
		log.Printf("[ExecutarRespostaRapida] ERRO: Formato de userID inválido, tipo: %T", userIDInterface)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Formato de ID de usuário inválido"})
		return
	}

	idParam := c.Param("id")
	log.Printf("[ExecutarRespostaRapida] ID da resposta recebido: %s", idParam)
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.Printf("[ExecutarRespostaRapida] ERRO: Falha ao fazer parse do ID da resposta: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	log.Printf("[ExecutarRespostaRapida] ID da resposta parseado: %s", id.String())

	// Ler o body bruto primeiro para debug
	bodyBytes, _ := c.GetRawData()
	log.Printf("[ExecutarRespostaRapida] Body recebido: %s", string(bodyBytes))
	
	// Recriar o body para o binding
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	var req struct {
		ChatID string `json:"chat_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ExecutarRespostaRapida] ERRO: Falha no binding JSON: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[ExecutarRespostaRapida] Dados bindados com sucesso - ChatID: %s", req.ChatID)

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
