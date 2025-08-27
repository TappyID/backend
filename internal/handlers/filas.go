package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"tappyone/internal/models"
	"gorm.io/gorm"
)

// FilasHandler gerencia as operações de filas
type FilasHandler struct {
	db *gorm.DB
}

// NewFilasHandler cria um novo handler para filas
func NewFilasHandler(db *gorm.DB) *FilasHandler {
	return &FilasHandler{db: db}
}

// ListarFilas - GET /api/filas
func (h *FilasHandler) ListarFilas(c *gin.Context) {
	var filas []models.Fila

	// Buscar filas com atendentes
	result := h.db.Preload("Atendentes.Usuario").Order("ordenacao ASC").Find(&filas)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar filas",
			"details": result.Error.Error(),
		})
		return
	}

	// Buscar estatísticas para cada fila
	var filasComStats []models.FilaWithStats
	for _, fila := range filas {
		stats := h.calcularEstatisticasFila(fila.ID)
		filaComStats := models.FilaWithStats{
			Fila:         fila,
			Estatisticas: stats,
		}
		filasComStats = append(filasComStats, filaComStats)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    filasComStats,
	})
}

// ObterFila - GET /api/filas/:id
func (h *FilasHandler) ObterFila(c *gin.Context) {
	id := c.Param("id")
	
	var fila models.Fila
	result := h.db.Preload("Atendentes.Usuario").First(&fila, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Fila não encontrada",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar fila",
			"details": result.Error.Error(),
		})
		return
	}

	// Calcular estatísticas
	stats := h.calcularEstatisticasFila(fila.ID)
	filaComStats := models.FilaWithStats{
		Fila:         fila,
		Estatisticas: stats,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    filaComStats,
	})
}

// CriarFila - POST /api/filas
func (h *FilasHandler) CriarFila(c *gin.Context) {
	var req models.FilaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos",
			"details": err.Error(),
		})
		return
	}

	// Verificar se a ordenação já existe
	var existingFila models.Fila
	result := h.db.Where("ordenacao = ?", req.Ordenacao).First(&existingFila)
	if result.Error == nil {
		// Ajustar ordenação das outras filas
		h.db.Model(&models.Fila{}).Where("ordenacao >= ?", req.Ordenacao).Update("ordenacao", gorm.Expr("ordenacao + 1"))
	}

	// Criar nova fila
	fila := models.Fila{
		Nome:          req.Nome,
		Descricao:     req.Descricao,
		Cor:           req.Cor,
		Ordenacao:     req.Ordenacao,
		Ativa:         req.Ativa,
		Prioridade:    req.Prioridade,
		ChatBot:       req.ChatBot,
		Kanban:        req.Kanban,
		WhatsappChats: req.WhatsappChats,
	}

	tx := h.db.Begin()
	if err := tx.Create(&fila).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao criar fila",
			"details": err.Error(),
		})
		return
	}

	// Associar atendentes à fila
	for _, atendenteID := range req.AtendentesIDs {
		filaAtendente := models.FilaAtendente{
			FilaID:    fila.ID,
			UsuarioID: atendenteID,
		}
		if err := tx.Create(&filaAtendente).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao associar atendentes à fila",
				"details": err.Error(),
			})
			return
		}
	}

	tx.Commit()

	// Retornar fila criada com atendentes
	h.db.Preload("Atendentes.Usuario").Where("id = ?", fila.ID).First(&fila)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    fila,
		"message": "Fila criada com sucesso",
	})
}

// AtualizarFila - PUT /api/filas/:id
func (h *FilasHandler) AtualizarFila(c *gin.Context) {
	id := c.Param("id")
	
	var req models.FilaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos",
			"details": err.Error(),
		})
		return
	}

	var fila models.Fila
	result := h.db.First(&fila, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Fila não encontrada",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar fila",
			"details": result.Error.Error(),
		})
		return
	}

	// Verificar se a ordenação mudou
	if req.Ordenacao != fila.Ordenacao {
		// Ajustar ordenação das outras filas
		if req.Ordenacao < fila.Ordenacao {
			h.db.Model(&models.Fila{}).Where("ordenacao >= ? AND ordenacao < ? AND id != ?", req.Ordenacao, fila.Ordenacao, id).Update("ordenacao", gorm.Expr("ordenacao + 1"))
		} else {
			h.db.Model(&models.Fila{}).Where("ordenacao > ? AND ordenacao <= ? AND id != ?", fila.Ordenacao, req.Ordenacao, id).Update("ordenacao", gorm.Expr("ordenacao - 1"))
		}
	}

	tx := h.db.Begin()

	// Atualizar dados da fila
	fila.Nome = req.Nome
	fila.Descricao = req.Descricao
	fila.Cor = req.Cor
	fila.Ordenacao = req.Ordenacao
	fila.Ativa = req.Ativa
	fila.Prioridade = req.Prioridade
	fila.ChatBot = req.ChatBot
	fila.Kanban = req.Kanban
	fila.WhatsappChats = req.WhatsappChats

	if err := tx.Save(&fila).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao atualizar fila",
			"details": err.Error(),
		})
		return
	}

	// Remover atendentes existentes
	if err := tx.Where("fila_id = ?", fila.ID).Delete(&models.FilaAtendente{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao atualizar atendentes",
			"details": err.Error(),
		})
		return
	}

	// Associar novos atendentes
	for _, atendenteID := range req.AtendentesIDs {
		filaAtendente := models.FilaAtendente{
			FilaID:    fila.ID,
			UsuarioID: atendenteID,
		}
		if err := tx.Create(&filaAtendente).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao associar atendentes à fila",
				"details": err.Error(),
			})
			return
		}
	}

	tx.Commit()

	// Retornar fila atualizada
	h.db.Preload("Atendentes.Usuario").First(&fila, fila.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    fila,
		"message": "Fila atualizada com sucesso",
	})
}

// DeletarFila - DELETE /api/filas/:id
func (h *FilasHandler) DeletarFila(c *gin.Context) {
	id := c.Param("id")
	
	var fila models.Fila
	result := h.db.First(&fila, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Fila não encontrada",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar fila",
			"details": result.Error.Error(),
		})
		return
	}

	tx := h.db.Begin()

	// Remover associações com atendentes
	if err := tx.Where("fila_id = ?", fila.ID).Delete(&models.FilaAtendente{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao remover associações da fila",
			"details": err.Error(),
		})
		return
	}

	// Deletar fila
	if err := tx.Delete(&fila).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao deletar fila",
			"details": err.Error(),
		})
		return
	}

	// Ajustar ordenação das filas restantes
	tx.Model(&models.Fila{}).Where("ordenacao > ?", fila.Ordenacao).Update("ordenacao", gorm.Expr("ordenacao - 1"))

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Fila deletada com sucesso",
	})
}

// ToggleFilaStatus - PATCH /api/filas/:id/toggle
func (h *FilasHandler) ToggleFilaStatus(c *gin.Context) {
	id := c.Param("id")
	
	var fila models.Fila
	result := h.db.First(&fila, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Fila não encontrada",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar fila",
			"details": result.Error.Error(),
		})
		return
	}

	// Toggle status
	fila.Ativa = !fila.Ativa
	if err := h.db.Save(&fila).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao atualizar status da fila",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    fila,
		"message": fmt.Sprintf("Fila %s com sucesso", map[bool]string{true: "ativada", false: "desativada"}[fila.Ativa]),
	})
}

// ReordenarFilas - POST /api/filas/reordenar
func (h *FilasHandler) ReordenarFilas(c *gin.Context) {
	var req struct {
		FilasOrdem []struct {
			ID        string `json:"id"`
			Ordenacao int    `json:"ordenacao"`
		} `json:"filasOrdem"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Dados inválidos",
			"details": err.Error(),
		})
		return
	}

	tx := h.db.Begin()

	for _, item := range req.FilasOrdem {
		if err := tx.Model(&models.Fila{}).Where("id = ?", item.ID).Update("ordenacao", item.Ordenacao).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao reordenar filas",
				"details": err.Error(),
			})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Filas reordenadas com sucesso",
	})
}

// calcularEstatisticasFila calcula as estatísticas de uma fila
func (h *FilasHandler) calcularEstatisticasFila(filaID string) models.FilaEstatisticas {
	stats := models.FilaEstatisticas{}

	// Por enquanto retornamos dados simulados
	// TODO: Implementar cálculos reais baseados nos atendimentos
	stats.TotalConversas = 0
	stats.ConversasAtivas = 0
	stats.TempoMedioResposta = 0.0
	stats.Satisfacao = 0.0

	return stats
}

// DuplicarFila - POST /api/filas/:id/duplicar
func (h *FilasHandler) DuplicarFila(c *gin.Context) {
	id := c.Param("id")
	
	var filaOriginal models.Fila
	result := h.db.Preload("Atendentes").First(&filaOriginal, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Fila não encontrada",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao buscar fila",
			"details": result.Error.Error(),
		})
		return
	}

	// Buscar maior ordenação atual
	var maxOrdenacao int
	h.db.Model(&models.Fila{}).Select("COALESCE(MAX(ordenacao), 0)").Scan(&maxOrdenacao)

	// Criar nova fila
	novaFila := models.Fila{
		Nome:          filaOriginal.Nome + " (Cópia)",
		Descricao:     filaOriginal.Descricao,
		Cor:           filaOriginal.Cor,
		Ordenacao:     maxOrdenacao + 1,
		Ativa:         false, // Nova fila começa inativa
		Prioridade:    filaOriginal.Prioridade,
		ChatBot:       filaOriginal.ChatBot,
		Kanban:        filaOriginal.Kanban,
		WhatsappChats: filaOriginal.WhatsappChats,
	}

	tx := h.db.Begin()
	if err := tx.Create(&novaFila).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Erro ao duplicar fila",
			"details": err.Error(),
		})
		return
	}

	// Duplicar atendentes
	for _, atendente := range filaOriginal.Atendentes {
		novoFilaAtendente := models.FilaAtendente{
			FilaID:    novaFila.ID,
			UsuarioID: atendente.UsuarioID,
		}
		if err := tx.Create(&novoFilaAtendente).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Erro ao duplicar atendentes da fila",
				"details": err.Error(),
			})
			return
		}
	}

	tx.Commit()

	// Retornar fila duplicada
	h.db.Preload("Atendentes.Usuario").First(&novaFila, novaFila.ID)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    novaFila,
		"message": "Fila duplicada com sucesso",
	})
}
