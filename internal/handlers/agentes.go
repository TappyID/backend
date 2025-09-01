package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tappyone/internal/models"
)

type AgentesHandler struct {
	db *gorm.DB
}

func NewAgentesHandler(db *gorm.DB) *AgentesHandler {
	return &AgentesHandler{db: db}
}

// GetAgentes busca todos os agentes do usuário autenticado
func (h *AgentesHandler) GetAgentes(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var agentes []models.AgenteIa
	
	// Query com paginação
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit
	
	search := c.Query("search")
	categoria := c.Query("categoria")
	
	query := h.db.Where("usuario_id = ?", userID)
	
	if search != "" {
		query = query.Where("nome ILIKE ? OR descricao ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	
	if categoria != "" {
		query = query.Where("categoria = ?", categoria)
	}
	
	var total int64
	query.Model(&models.AgenteIa{}).Count(&total)
	
	err := query.Limit(limit).Offset(offset).Order("nome ASC").Find(&agentes).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agents"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agentes": agentes,
		"total":   total,
		"page":    page,
		"limit":   limit,
	})
}

// GetAgente busca um agente específico por ID
func (h *AgentesHandler) GetAgente(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	agenteID := c.Param("id")
	var agente models.AgenteIa

	err := h.db.Where("id = ? AND usuario_id = ?", agenteID, userID).First(&agente).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agent"})
		return
	}

	c.JSON(http.StatusOK, agente)
}

// CreateAgente cria um novo agente de IA
func (h *AgentesHandler) CreateAgente(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		Nome      string  `json:"nome" binding:"required"`
		Descricao *string `json:"descricao"`
		Prompt    string  `json:"prompt" binding:"required"`
		Modelo    string  `json:"modelo" binding:"required"`
		Categoria *string `json:"categoria"`
		Funcao    *string `json:"funcao"`
		Nicho     *string `json:"nicho"`
		Ativo     *bool   `json:"ativo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agente := models.AgenteIa{
		UsuarioID: userID.(string),
		Nome:      req.Nome,
		Descricao: req.Descricao,
		Prompt:    req.Prompt,
		Modelo:    req.Modelo,
		Categoria: req.Categoria,
		Funcao:    req.Funcao,
		Nicho:     req.Nicho,
		Ativo:     true,
	}

	if req.Ativo != nil {
		agente.Ativo = *req.Ativo
	}

	if err := h.db.Create(&agente).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agent"})
		return
	}

	c.JSON(http.StatusCreated, agente)
}

// UpdateAgente atualiza um agente existente
func (h *AgentesHandler) UpdateAgente(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	agenteID := c.Param("id")
	
	var agente models.AgenteIa
	err := h.db.Where("id = ? AND usuario_id = ?", agenteID, userID).First(&agente).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agent"})
		return
	}

	var req struct {
		Nome      *string `json:"nome"`
		Descricao *string `json:"descricao"`
		Prompt    *string `json:"prompt"`
		Modelo    *string `json:"modelo"`
		Categoria *string `json:"categoria"`
		Funcao    *string `json:"funcao"`
		Nicho     *string `json:"nicho"`
		Ativo     *bool   `json:"ativo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Nome != nil {
		updates["nome"] = *req.Nome
	}
	if req.Descricao != nil {
		updates["descricao"] = *req.Descricao
	}
	if req.Prompt != nil {
		updates["prompt"] = *req.Prompt
	}
	if req.Modelo != nil {
		updates["modelo"] = *req.Modelo
	}
	if req.Categoria != nil {
		updates["categoria"] = *req.Categoria
	}
	if req.Funcao != nil {
		updates["funcao"] = *req.Funcao
	}
	if req.Nicho != nil {
		updates["nicho"] = *req.Nicho
	}
	if req.Ativo != nil {
		updates["ativo"] = *req.Ativo
	}

	if err := h.db.Model(&agente).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update agent"})
		return
	}

	// Recarregar o agente atualizado
	h.db.Where("id = ?", agenteID).First(&agente)

	c.JSON(http.StatusOK, agente)
}

// DeleteAgente remove um agente
func (h *AgentesHandler) DeleteAgente(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	agenteID := c.Param("id")
	
	var agente models.AgenteIa
	err := h.db.Where("id = ? AND usuario_id = ?", agenteID, userID).First(&agente).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agent"})
		return
	}

	// Primeiro remove todas as ativações deste agente
	h.db.Where("agente_id = ?", agenteID).Delete(&models.ChatAgente{})
	
	// Remove o agente
	if err := h.db.Delete(&agente).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete agent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agent deleted successfully"})
}

// ToggleAgente alterna o status ativo/inativo de um agente
func (h *AgentesHandler) ToggleAgente(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	agenteID := c.Param("id")
	
	var agente models.AgenteIa
	err := h.db.Where("id = ? AND usuario_id = ?", agenteID, userID).First(&agente).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agent"})
		return
	}

	// Toggle status
	agente.Ativo = !agente.Ativo
	
	if err := h.db.Save(&agente).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle agent status"})
		return
	}

	c.JSON(http.StatusOK, agente)
}

// GetChatAgente verifica se um chat tem agente ativo
func (h *AgentesHandler) GetChatAgente(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	chatID := c.Param("chatId")
	
	var chatAgente models.ChatAgente
	err := h.db.Preload("Agente").Where("chat_id = ? AND usuario_id = ? AND ativo = true", chatID, userID).First(&chatAgente).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"ativo": false,
				"agente": nil,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chat agent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ativo": true,
		"agente": chatAgente.Agente,
		"chatAgente": chatAgente,
	})
}

// ActivateAgentForChat ativa um agente para um chat específico
func (h *AgentesHandler) ActivateAgentForChat(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	chatID := c.Param("chatId")
	
	var req struct {
		AgenteID string `json:"agenteId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar se o agente existe e pertence ao usuário
	var agente models.AgenteIa
	err := h.db.Where("id = ? AND usuario_id = ? AND ativo = true", req.AgenteID, userID).First(&agente).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Agent not found or inactive"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch agent"})
		return
	}

	// Desativar agente anterior se existir
	h.db.Model(&models.ChatAgente{}).Where("chat_id = ? AND usuario_id = ?", chatID, userID).Update("ativo", false)
	
	// Verificar se já existe uma ativação para este agente e chat
	var chatAgente models.ChatAgente
	err = h.db.Where("chat_id = ? AND agente_id = ? AND usuario_id = ?", chatID, req.AgenteID, userID).First(&chatAgente).Error
	
	if err == gorm.ErrRecordNotFound {
		// Criar nova ativação
		chatAgente = models.ChatAgente{
			ChatID:    chatID,
			AgenteID:  req.AgenteID,
			UsuarioID: userID.(string),
			Ativo:     true,
		}
		
		if err := h.db.Create(&chatAgente).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate agent"})
			return
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing activation"})
		return
	} else {
		// Reativar existente
		chatAgente.Ativo = true
		if err := h.db.Save(&chatAgente).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate agent"})
			return
		}
	}

	// Carregar dados completos para resposta
	h.db.Preload("Agente").Where("id = ?", chatAgente.ID).First(&chatAgente)

	c.JSON(http.StatusOK, gin.H{
		"message": "Agent activated successfully",
		"chatAgente": chatAgente,
	})
}

// DeactivateAgentForChat desativa agente para um chat específico
func (h *AgentesHandler) DeactivateAgentForChat(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	chatID := c.Param("chatId")
	
	result := h.db.Model(&models.ChatAgente{}).Where("chat_id = ? AND usuario_id = ?", chatID, userID).Update("ativo", false)
	
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate agent"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Agent deactivated successfully",
	})
}

// GetAgentesAtivos retorna lista de agentes ativos do usuário para seleção
func (h *AgentesHandler) GetAgentesAtivos(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var agentes []models.AgenteIa
	
	err := h.db.Where("usuario_id = ? AND ativo = true", userID).Select("id, nome, descricao, categoria, funcao").Order("nome ASC").Find(&agentes).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch active agents"})
		return
	}

	c.JSON(http.StatusOK, agentes)
}
