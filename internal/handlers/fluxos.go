package handlers

import (
	"net/http"
	"strconv"
	"tappyone/internal/models"
	"tappyone/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FluxosHandler struct {
	DB                    *gorm.DB
	FluxoExecutionService *services.FluxoExecutionService
}

func NewFluxosHandler(db *gorm.DB, fluxoExecutionService *services.FluxoExecutionService) *FluxosHandler {
	return &FluxosHandler{
		DB:                    db,
		FluxoExecutionService: fluxoExecutionService,
	}
}

// ListFluxos - GET /api/fluxos
func (h *FluxosHandler) ListFluxos(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var fluxos []models.Fluxo
	
	// Query com filtros opcionais
	query := h.DB.Preload("Quadro").Preload("Nos")
	
	// Filtrar por quadro se especificado
	if quadroID := c.Query("quadro_id"); quadroID != "" {
		query = query.Where("quadro_id = ?", quadroID)
	}
	
	// Filtrar por status ativo
	if ativo := c.Query("ativo"); ativo != "" {
		if ativo == "true" {
			query = query.Where("ativo = ?", true)
		} else if ativo == "false" {
			query = query.Where("ativo = ?", false)
		}
	}

	// Buscar fluxos onde o usuário é dono do quadro
	if err := query.Joins("JOIN quadros ON fluxos.quadro_id = quadros.id").
		Where("quadros.usuario_id = ?", userID).
		Find(&fluxos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar fluxos"})
		return
	}

	c.JSON(http.StatusOK, fluxos)
}

// GetFluxo - GET /api/fluxos/:id
func (h *FluxosHandler) GetFluxo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	fluxoID := c.Param("id")
	var fluxo models.Fluxo

	// Buscar fluxo com relacionamentos
	if err := h.DB.Preload("Quadro").
		Preload("Nos").
		Preload("Nos.ConexoesDe").
		Preload("Nos.ConexoesPara").
		Joins("JOIN quadros ON fluxos.quadro_id = quadros.id").
		Where("fluxos.id = ? AND quadros.usuario_id = ?", fluxoID, userID).
		First(&fluxo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Fluxo não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar fluxo"})
		return
	}

	c.JSON(http.StatusOK, fluxo)
}

// CreateFluxo - POST /api/fluxos
func (h *FluxosHandler) CreateFluxo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		Nome      string `json:"nome" binding:"required"`
		Descricao string `json:"descricao"`
		QuadroID  string `json:"quadroId" binding:"required"`
		Ativo     *bool  `json:"ativo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar se o quadro pertence ao usuário
	var quadro models.Quadro
	if err := h.DB.Where("id = ? AND usuario_id = ?", req.QuadroID, userID).
		First(&quadro).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Quadro não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar quadro"})
		return
	}

	// Criar fluxo
	fluxo := models.Fluxo{
		Nome:     req.Nome,
		QuadroID: req.QuadroID,
		Ativo:    true,
	}

	if req.Descricao != "" {
		fluxo.Descricao = &req.Descricao
	}

	if req.Ativo != nil {
		fluxo.Ativo = *req.Ativo
	}

	if err := h.DB.Create(&fluxo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar fluxo"})
		return
	}

	// Recarregar com relacionamentos
	h.DB.Preload("Quadro").Preload("Nos").First(&fluxo, fluxo.ID)

	c.JSON(http.StatusCreated, fluxo)
}

// UpdateFluxo - PUT /api/fluxos/:id
func (h *FluxosHandler) UpdateFluxo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	fluxoID := c.Param("id")
	
	var req struct {
		Nome      string `json:"nome"`
		Descricao string `json:"descricao"`
		Ativo     *bool  `json:"ativo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Buscar fluxo
	var fluxo models.Fluxo
	if err := h.DB.Joins("JOIN quadros ON fluxos.quadro_id = quadros.id").
		Where("fluxos.id = ? AND quadros.usuario_id = ?", fluxoID, userID).
		First(&fluxo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Fluxo não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar fluxo"})
		return
	}

	// Atualizar campos
	if req.Nome != "" {
		fluxo.Nome = req.Nome
	}
	
	if req.Descricao != "" {
		fluxo.Descricao = &req.Descricao
	}
	
	if req.Ativo != nil {
		fluxo.Ativo = *req.Ativo
	}

	if err := h.DB.Save(&fluxo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar fluxo"})
		return
	}

	// Recarregar com relacionamentos
	h.DB.Preload("Quadro").Preload("Nos").First(&fluxo, fluxo.ID)

	c.JSON(http.StatusOK, fluxo)
}

// DeleteFluxo - DELETE /api/fluxos/:id
func (h *FluxosHandler) DeleteFluxo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	fluxoID := c.Param("id")

	// Buscar fluxo para verificar propriedade
	var fluxo models.Fluxo
	if err := h.DB.Joins("JOIN quadros ON fluxos.quadro_id = quadros.id").
		Where("fluxos.id = ? AND quadros.usuario_id = ?", fluxoID, userID).
		First(&fluxo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Fluxo não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar fluxo"})
		return
	}

	// Deletar fluxo (cascade delete vai remover nós e conexões)
	if err := h.DB.Delete(&fluxo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar fluxo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fluxo deletado com sucesso"})
}

// ToggleFluxo - PUT /api/fluxos/:id/toggle
func (h *FluxosHandler) ToggleFluxo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	fluxoID := c.Param("id")

	// Buscar fluxo
	var fluxo models.Fluxo
	if err := h.DB.Joins("JOIN quadros ON fluxos.quadro_id = quadros.id").
		Where("fluxos.id = ? AND quadros.usuario_id = ?", fluxoID, userID).
		First(&fluxo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Fluxo não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar fluxo"})
		return
	}

	// Toggle status
	fluxo.Ativo = !fluxo.Ativo

	if err := h.DB.Save(&fluxo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar fluxo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    fluxo.ID,
		"ativo": fluxo.Ativo,
		"message": "Status do fluxo atualizado",
	})
}

// CreateFluxoNo - POST /api/fluxos/:id/nos
func (h *FluxosHandler) CreateFluxoNo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	fluxoID := c.Param("id")
	
	var req struct {
		Nome         string      `json:"nome" binding:"required"`
		Tipo         string      `json:"tipo" binding:"required"`
		Configuracao interface{} `json:"configuracao"`
		Posicao      interface{} `json:"posicao"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar se fluxo existe e pertence ao usuário
	var fluxo models.Fluxo
	if err := h.DB.Joins("JOIN quadros ON fluxos.quadro_id = quadros.id").
		Where("fluxos.id = ? AND quadros.usuario_id = ?", fluxoID, userID).
		First(&fluxo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Fluxo não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar fluxo"})
		return
	}

	// Criar nó
	no := models.FluxoNo{
		Nome:    req.Nome,
		Tipo:    req.Tipo,
		FluxoID: fluxoID,
	}

	// Converter configuração e posição para JSONB
	if req.Configuracao != nil {
		if configMap, ok := req.Configuracao.(map[string]interface{}); ok {
			no.Configuracao = models.JSONB(configMap)
		}
	}
	if req.Posicao != nil {
		if posMap, ok := req.Posicao.(map[string]interface{}); ok {
			no.Posicao = models.JSONB(posMap)
		}
	}

	if err := h.DB.Create(&no).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar nó"})
		return
	}

	c.JSON(http.StatusCreated, no)
}

// UpdateFluxoNo - PUT /api/fluxos/:fluxoId/nos/:noId
func (h *FluxosHandler) UpdateFluxoNo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	fluxoID := c.Param("fluxoId")
	noID := c.Param("noId")
	
	var req struct {
		Nome         string      `json:"nome"`
		Tipo         string      `json:"tipo"`
		Configuracao interface{} `json:"configuracao"`
		Posicao      interface{} `json:"posicao"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar se nó existe e pertence ao usuário
	var no models.FluxoNo
	if err := h.DB.Joins("JOIN fluxos ON fluxo_nos.fluxo_id = fluxos.id").
		Joins("JOIN quadros ON fluxos.quadro_id = quadros.id").
		Where("fluxo_nos.id = ? AND fluxos.id = ? AND quadros.usuario_id = ?", noID, fluxoID, userID).
		First(&no).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Nó não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar nó"})
		return
	}

	// Atualizar campos
	if req.Nome != "" {
		no.Nome = req.Nome
	}
	if req.Tipo != "" {
		no.Tipo = req.Tipo
	}
	if req.Configuracao != nil {
		if configMap, ok := req.Configuracao.(map[string]interface{}); ok {
			no.Configuracao = models.JSONB(configMap)
		}
	}
	if req.Posicao != nil {
		if posMap, ok := req.Posicao.(map[string]interface{}); ok {
			no.Posicao = models.JSONB(posMap)
		}
	}

	if err := h.DB.Save(&no).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar nó"})
		return
	}

	c.JSON(http.StatusOK, no)
}

// DeleteFluxoNo - DELETE /api/fluxos/:fluxoId/nos/:noId
func (h *FluxosHandler) DeleteFluxoNo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	fluxoID := c.Param("fluxoId")
	noID := c.Param("noId")

	// Verificar se nó existe e pertence ao usuário
	var no models.FluxoNo
	if err := h.DB.Joins("JOIN fluxos ON fluxo_nos.fluxo_id = fluxos.id").
		Joins("JOIN quadros ON fluxos.quadro_id = quadros.id").
		Where("fluxo_nos.id = ? AND fluxos.id = ? AND quadros.usuario_id = ?", noID, fluxoID, userID).
		First(&no).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Nó não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar nó"})
		return
	}

	// Deletar nó (cascade delete vai remover conexões)
	if err := h.DB.Delete(&no).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar nó"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Nó deletado com sucesso"})
}

// CreateFluxoConexao - POST /api/fluxos/:id/conexoes
func (h *FluxosHandler) CreateFluxoConexao(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	fluxoID := c.Param("id")
	
	var req struct {
		DeID   string `json:"deId" binding:"required"`
		ParaID string `json:"paraId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar se os nós existem e pertencem ao fluxo
	var count int64
	if err := h.DB.Table("fluxo_nos").
		Joins("JOIN fluxos ON fluxo_nos.fluxo_id = fluxos.id").
		Joins("JOIN quadros ON fluxos.quadro_id = quadros.id").
		Where("fluxo_nos.id IN (?, ?) AND fluxos.id = ? AND quadros.usuario_id = ?", 
			req.DeID, req.ParaID, fluxoID, userID).
		Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar nós"})
		return
	}

	if count != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nós não encontrados ou não pertencem ao fluxo"})
		return
	}

	// Verificar se conexão já existe
	var existingCount int64
	h.DB.Model(&models.FluxoConexao{}).
		Where("de_id = ? AND para_id = ?", req.DeID, req.ParaID).
		Count(&existingCount)

	if existingCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conexão já existe"})
		return
	}

	// Criar conexão
	conexao := models.FluxoConexao{
		DeID:   req.DeID,
		ParaID: req.ParaID,
	}

	if err := h.DB.Create(&conexao).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar conexão"})
		return
	}

	// Recarregar com relacionamentos
	h.DB.Preload("De").Preload("Para").First(&conexao, conexao.ID)

	c.JSON(http.StatusCreated, conexao)
}

// DeleteFluxoConexao - DELETE /api/fluxos/:fluxoId/conexoes/:conexaoId
func (h *FluxosHandler) DeleteFluxoConexao(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	fluxoID := c.Param("fluxoId")
	conexaoID := c.Param("conexaoId")

	// Verificar se conexão existe e pertence ao usuário
	var conexao models.FluxoConexao
	if err := h.DB.Joins("JOIN fluxo_nos de ON fluxo_conexoes.de_id = de.id").
		Joins("JOIN fluxos ON de.fluxo_id = fluxos.id").
		Joins("JOIN quadros ON fluxos.quadro_id = quadros.id").
		Where("fluxo_conexoes.id = ? AND fluxos.id = ? AND quadros.usuario_id = ?", 
			conexaoID, fluxoID, userID).
		First(&conexao).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Conexão não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar conexão"})
		return
	}

	// Deletar conexão
	if err := h.DB.Delete(&conexao).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar conexão"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conexão deletada com sucesso"})
}

// GetFluxosStats - GET /api/fluxos/stats
func (h *FluxosHandler) GetFluxosStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var stats struct {
		TotalFluxos  int64 `json:"totalFluxos"`
		FluxosAtivos int64 `json:"fluxosAtivos"`
		TotalNos     int64 `json:"totalNos"`
		TotalConexoes int64 `json:"totalConexoes"`
	}

	// Total de fluxos do usuário
	h.DB.Model(&models.Fluxo{}).
		Joins("JOIN quadros ON fluxos.quadro_id = quadros.id").
		Where("quadros.usuario_id = ?", userID).
		Count(&stats.TotalFluxos)

	// Fluxos ativos
	h.DB.Model(&models.Fluxo{}).
		Joins("JOIN quadros ON fluxos.quadro_id = quadros.id").
		Where("quadros.usuario_id = ? AND fluxos.ativo = ?", userID, true).
		Count(&stats.FluxosAtivos)

	// Total de nós
	h.DB.Model(&models.FluxoNo{}).
		Joins("JOIN fluxos ON fluxo_nos.fluxo_id = fluxos.id").
		Joins("JOIN quadros ON fluxos.quadro_id = quadros.id").
		Where("quadros.usuario_id = ?", userID).
		Count(&stats.TotalNos)

	// Total de conexões
	h.DB.Model(&models.FluxoConexao{}).
		Joins("JOIN fluxo_nos de ON fluxo_conexoes.de_id = de.id").
		Joins("JOIN fluxos ON de.fluxo_id = fluxos.id").
		Joins("JOIN quadros ON fluxos.quadro_id = quadros.id").
		Where("quadros.usuario_id = ?", userID).
		Count(&stats.TotalConexoes)

	c.JSON(http.StatusOK, stats)
}

// ExecuteFluxo - POST /api/fluxos/:id/execute
func (h *FluxosHandler) ExecuteFluxo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	fluxoID := c.Param("id")
	
	var req struct {
		ContatoID string                 `json:"contatoId"`
		Dados     map[string]interface{} `json:"dados"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Buscar fluxo
	var fluxo models.Fluxo
	if err := h.DB.Preload("Nos").
		Joins("JOIN quadros ON fluxos.quadro_id = quadros.id").
		Where("fluxos.id = ? AND quadros.usuario_id = ?", fluxoID, userID).
		First(&fluxo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Fluxo não encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar fluxo"})
		return
	}

	if !fluxo.Ativo {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Fluxo não está ativo"})
		return
	}

	// TODO: Implementar lógica de execução do fluxo
	// Por enquanto, retornamos sucesso
	c.JSON(http.StatusOK, gin.H{
		"message":   "Fluxo executado com sucesso",
		"fluxoId":   fluxoID,
		"contatoId": req.ContatoID,
		"timestamp": strconv.FormatInt(c.GetInt64("timestamp"), 10),
	})
}
