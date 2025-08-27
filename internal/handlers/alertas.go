package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tappyone/internal/models"
	"tappyone/internal/services"
)

type AlertasHandler struct {
	db      *gorm.DB
	authSvc *services.AuthService
}

func NewAlertasHandler(db *gorm.DB, authSvc *services.AuthService) *AlertasHandler {
	return &AlertasHandler{
		db:      db,
		authSvc: authSvc,
	}
}

// ListarAlertas - GET /api/alertas
func (h *AlertasHandler) ListarAlertas(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Usuário não autenticado",
		})
		return
	}

	var alertas []models.Alerta

	// Buscar alertas do usuário ordenados por criação
	if err := h.db.Where("user_id = ?", userID).
		Order("criado_em DESC").
		Find(&alertas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Erro ao buscar alertas",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    alertas,
	})
}

// CriarAlerta - POST /api/alertas
func (h *AlertasHandler) CriarAlerta(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Usuário não autenticado",
		})
		return
	}

	var alerta models.Alerta
	if err := c.ShouldBindJSON(&alerta); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Dados inválidos",
			"details": err.Error(),
		})
		return
	}

	// Validações
	if alerta.Titulo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Título é obrigatório",
		})
		return
	}

	// Definir usuário
	alerta.UserID = userID.(string)

	// Inicializar estatísticas
	alerta.Estatisticas = models.EstatisticasAlerta{
		TotalDisparos: 0,
		DisparosHoje:  0,
		TaxaResolucao: 0,
	}

	// Criar alerta
	if err := h.db.Create(&alerta).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Erro ao criar alerta",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    alerta,
		"message": "Alerta criado com sucesso",
	})
}

// ObterAlerta - GET /api/alertas/:id
func (h *AlertasHandler) ObterAlerta(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Usuário não autenticado",
		})
		return
	}

	id := c.Param("id")
	var alerta models.Alerta

	if err := h.db.Where("id = ? AND user_id = ?", id, userID).
		First(&alerta).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Alerta não encontrado",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Erro ao buscar alerta",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    alerta,
	})
}

// AtualizarAlerta - PUT /api/alertas/:id
func (h *AlertasHandler) AtualizarAlerta(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Usuário não autenticado",
		})
		return
	}

	id := c.Param("id")
	var alerta models.Alerta

	// Buscar alerta existente
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).
		First(&alerta).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Alerta não encontrado",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Erro ao buscar alerta",
				"details": err.Error(),
			})
		}
		return
	}

	var dadosAtualizacao models.Alerta
	if err := c.ShouldBindJSON(&dadosAtualizacao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Dados inválidos",
			"details": err.Error(),
		})
		return
	}

	// Atualizar campos permitidos
	if dadosAtualizacao.Titulo != "" {
		alerta.Titulo = dadosAtualizacao.Titulo
	}
	if dadosAtualizacao.Descricao != "" {
		alerta.Descricao = dadosAtualizacao.Descricao
	}
	if dadosAtualizacao.Tipo != "" {
		alerta.Tipo = dadosAtualizacao.Tipo
	}
	if dadosAtualizacao.Prioridade != "" {
		alerta.Prioridade = dadosAtualizacao.Prioridade
	}
	if dadosAtualizacao.Status != "" {
		alerta.Status = dadosAtualizacao.Status
	}
	if dadosAtualizacao.Cor != "" {
		alerta.Cor = dadosAtualizacao.Cor
	}
	if dadosAtualizacao.Icone != "" {
		alerta.Icone = dadosAtualizacao.Icone
	}
	
	// Atualizar configurações se fornecidas
	if len(dadosAtualizacao.Configuracoes.Destinatarios) > 0 {
		alerta.Configuracoes = dadosAtualizacao.Configuracoes
	}

	// Salvar alterações
	if err := h.db.Save(&alerta).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Erro ao atualizar alerta",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    alerta,
		"message": "Alerta atualizado com sucesso",
	})
}

// DeletarAlerta - DELETE /api/alertas/:id
func (h *AlertasHandler) DeletarAlerta(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Usuário não autenticado",
		})
		return
	}

	id := c.Param("id")
	var alerta models.Alerta

	// Buscar alerta existente
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).
		First(&alerta).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Alerta não encontrado",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Erro ao buscar alerta",
				"details": err.Error(),
			})
		}
		return
	}

	// Deletar histórico relacionado
	h.db.Where("alerta_id = ?", id).Delete(&models.HistoricoAlerta{})

	// Deletar alerta
	if err := h.db.Delete(&alerta).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Erro ao deletar alerta",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Alerta deletado com sucesso",
	})
}

// AlternarStatusAlerta - PATCH /api/alertas/:id/status
func (h *AlertasHandler) AlternarStatusAlerta(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Usuário não autenticado",
		})
		return
	}

	id := c.Param("id")
	var alerta models.Alerta

	// Buscar alerta existente
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).
		First(&alerta).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Alerta não encontrado",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Erro ao buscar alerta",
				"details": err.Error(),
			})
		}
		return
	}

	var request struct {
		Status models.StatusAlerta `json:"status"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Status inválido",
			"details": err.Error(),
		})
		return
	}

	// Atualizar status
	alerta.Status = request.Status
	if err := h.db.Save(&alerta).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Erro ao atualizar status do alerta",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    alerta,
		"message": "Status do alerta atualizado com sucesso",
	})
}

// ObterHistoricoAlerta - GET /api/alertas/:id/historico
func (h *AlertasHandler) ObterHistoricoAlerta(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Usuário não autenticado",
		})
		return
	}

	id := c.Param("id")
	
	// Verificar se alerta pertence ao usuário
	var alerta models.Alerta
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).
		First(&alerta).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Alerta não encontrado",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Erro ao buscar alerta",
				"details": err.Error(),
			})
		}
		return
	}

	// Buscar histórico
	var historico []models.HistoricoAlerta
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	if err := h.db.Where("alerta_id = ?", id).
		Order("criado_em DESC").
		Limit(limit).
		Find(&historico).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Erro ao buscar histórico do alerta",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    historico,
	})
}

// ObterEstatisticasAlertas - GET /api/alertas/stats
func (h *AlertasHandler) ObterEstatisticasAlertas(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Usuário não autenticado",
		})
		return
	}

	// Estatísticas básicas
	var stats struct {
		TotalAlertas    int64 `json:"totalAlertas"`
		AlertasAtivos   int64 `json:"alertasAtivos"`
		AlertasPausados int64 `json:"alertasPausados"`
		AlertasCriticos int64 `json:"alertasCriticos"`
	}

	// Total de alertas
	h.db.Model(&models.Alerta{}).Where("user_id = ?", userID).Count(&stats.TotalAlertas)
	
	// Alertas ativos
	h.db.Model(&models.Alerta{}).Where("user_id = ? AND status = ?", userID, "ativo").Count(&stats.AlertasAtivos)
	
	// Alertas pausados
	h.db.Model(&models.Alerta{}).Where("user_id = ? AND status = ?", userID, "pausado").Count(&stats.AlertasPausados)
	
	// Alertas críticos
	h.db.Model(&models.Alerta{}).Where("user_id = ? AND prioridade = ?", userID, "critica").Count(&stats.AlertasCriticos)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}
