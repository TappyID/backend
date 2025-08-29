package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"tappyone/internal/models"
	"gorm.io/gorm"
)

// AssinaturasHandler gerencia as assinaturas
type AssinaturasHandler struct {
	db *gorm.DB
}

func NewAssinaturasHandler(db *gorm.DB) *AssinaturasHandler {
	return &AssinaturasHandler{db: db}
}

// ListAssinaturas lista todas as assinaturas do usuário
func (h *AssinaturasHandler) ListAssinaturas(c *gin.Context) {
	// Verificar autenticação
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr := userID.(string)
	contatoJID := c.Query("contato_id")
	status := c.Query("status")

	var assinaturas []models.Assinatura
	query := h.db.Where("usuario_id = ?", userIDStr)
	
	if contatoJID != "" {
		// Converter JID para número de telefone e buscar contato
		numeroTelefone := strings.Replace(contatoJID, "@c.us", "", 1)
		
		// Buscar contato pelo número de telefone
		var contato models.Contato
		if err := h.db.Where("numero_telefone = ?", numeroTelefone).First(&contato).Error; err == nil {
			// Se contato encontrado, filtrar por UUID do contato
			query = query.Where("contato_id = ?", contato.ID)
		} else {
			// Se contato não encontrado, não retornar nenhuma assinatura
			c.JSON(http.StatusOK, []models.Assinatura{})
			return
		}
	}
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if err := query.Order("criado_em DESC").Find(&assinaturas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar assinaturas"})
		return
	}

	c.JSON(http.StatusOK, assinaturas)
}

// GetAssinatura busca uma assinatura específica
func (h *AssinaturasHandler) GetAssinatura(c *gin.Context) {
	userID := c.GetString("user_id")
	assinaturaID := c.Param("id")

	var assinatura models.Assinatura
	if err := h.db.Where("id = ? AND usuario_id = ?", assinaturaID, userID).First(&assinatura).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assinatura não encontrada"})
		return
	}

	c.JSON(http.StatusOK, assinatura)
}

// CreateAssinatura cria uma nova assinatura
func (h *AssinaturasHandler) CreateAssinatura(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		Nome            string     `json:"nome" binding:"required"`
		Plano           string     `json:"plano" binding:"required"`
		FormaPagamento  string     `json:"forma_pagamento" binding:"required"`
		LinkPagamento   *string    `json:"link_pagamento"`
		Valor           float64    `json:"valor" binding:"required"`
		Renovacao       string     `json:"renovacao" binding:"required"`
		DataInicio      *time.Time `json:"data_inicio"`
		DataFim         *time.Time `json:"data_fim"`
		ContatoID       string     `json:"contato_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Extrair número de telefone do JID (remove @c.us)
	numeroTelefone := strings.Replace(req.ContatoID, "@c.us", "", 1)
	
	// Buscar ou criar contato baseado no número de telefone
	var contato models.Contato
	err := h.db.Where("numero_telefone = ?", numeroTelefone).First(&contato).Error
	if err != nil {
		// Buscar ou criar uma sessão WhatsApp padrão
		var sessaoWhatsapp models.SessaoWhatsApp
		err := h.db.Where("nome_sessao = ?", "default").First(&sessaoWhatsapp).Error
		if err != nil {
			// Criar sessão padrão se não existir
			sessaoWhatsapp = models.SessaoWhatsApp{
				NomeSessao: "default",
				Status:     "ATIVO",
				UsuarioID:  userID,
			}
			if err := h.db.Create(&sessaoWhatsapp).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar sessão padrão"})
				return
			}
		}
		
		// Se contato não existe, criar um novo
		contato = models.Contato{
			NumeroTelefone:   numeroTelefone,
			Nome:             &numeroTelefone, // Usar número como nome temporário
			SessaoWhatsappID: sessaoWhatsapp.ID, // Usar sessão padrão
		}
		if err := h.db.Create(&contato).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar contato"})
			return
		}
	}

	// Definir data de início (usar data atual se não fornecida)
	var dataInicio time.Time
	if req.DataInicio != nil {
		dataInicio = *req.DataInicio
	} else {
		dataInicio = time.Now()
	}

	// Buscar ou criar plano baseado no nome
	var plano models.Plano
	err = h.db.Where("nome = ?", req.Plano).First(&plano).Error
	if err != nil {
		// Se plano não existe, criar um novo
		plano = models.Plano{
			Nome:      req.Plano,
			Descricao: nil,
			Preco:     req.Valor,
			Intervalo: req.Renovacao,
			Ativo:     true,
		}
		if err := h.db.Create(&plano).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar plano"})
			return
		}
	}

	assinatura := models.Assinatura{
		Nome:           req.Nome,
		FormaPagamento: models.TipoCobranca(req.FormaPagamento),
		LinkPagamento:  req.LinkPagamento,
		Valor:          req.Valor,
		Renovacao:      req.Renovacao,
		DataInicio:     dataInicio,
		DataFim:        req.DataFim,
		Status:         "ATIVA",
		UsuarioID:      userID,
		ContatoID:      contato.ID, // Usar ID do contato encontrado/criado
		PlanoID:       plano.ID,   // Usar ID do plano encontrado/criado
	}

	if err := h.db.Create(&assinatura).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar assinatura"})
		return
	}

	c.JSON(http.StatusCreated, assinatura)
}

// UpdateAssinatura atualiza uma assinatura existente
func (h *AssinaturasHandler) UpdateAssinatura(c *gin.Context) {
	userID := c.GetString("user_id")
	assinaturaID := c.Param("id")

	var req struct {
		Nome           *string    `json:"nome"`
		FormaPagamento *string    `json:"forma_pagamento"`
		LinkPagamento  *string    `json:"link_pagamento"`
		Valor          *float64   `json:"valor"`
		Renovacao      *string    `json:"renovacao"`
		DataInicio     *time.Time `json:"data_inicio"`
		DataFim        *time.Time `json:"data_fim"`
		Status         *string    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var assinatura models.Assinatura
	if err := h.db.Where("id = ? AND usuario_id = ?", assinaturaID, userID).First(&assinatura).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assinatura não encontrada"})
		return
	}

	// Atualizar campos fornecidos
	updates := make(map[string]interface{})
	if req.Nome != nil {
		updates["nome"] = *req.Nome
	}
	if req.FormaPagamento != nil {
		updates["forma_pagamento"] = *req.FormaPagamento
	}
	if req.LinkPagamento != nil {
		updates["link_pagamento"] = *req.LinkPagamento
	}
	if req.Valor != nil {
		updates["valor"] = *req.Valor
	}
	if req.Renovacao != nil {
		updates["renovacao"] = *req.Renovacao
	}
	if req.DataInicio != nil {
		updates["data_inicio"] = *req.DataInicio
	}
	if req.DataFim != nil {
		updates["data_fim"] = *req.DataFim
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	updates["atualizado_em"] = time.Now()

	if err := h.db.Model(&assinatura).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar assinatura"})
		return
	}

	// Recarregar dados atualizados
	if err := h.db.Where("id = ?", assinaturaID).First(&assinatura).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao recarregar assinatura"})
		return
	}

	c.JSON(http.StatusOK, assinatura)
}

// DeleteAssinatura exclui uma assinatura
func (h *AssinaturasHandler) DeleteAssinatura(c *gin.Context) {
	userID := c.GetString("user_id")
	assinaturaID := c.Param("id")

	var assinatura models.Assinatura
	if err := h.db.Where("id = ? AND usuario_id = ?", assinaturaID, userID).First(&assinatura).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assinatura não encontrada"})
		return
	}

	if err := h.db.Delete(&assinatura).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir assinatura"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Assinatura excluída com sucesso"})
}

// GetAssinaturasByContato busca todas as assinaturas de um contato específico
func (h *AssinaturasHandler) GetAssinaturasByContato(c *gin.Context) {
	userID := c.GetString("user_id")
	contatoID := c.Param("contato_id")

	var assinaturas []models.Assinatura
	if err := h.db.Where("usuario_id = ? AND contato_id = ?", userID, contatoID).Order("criado_em DESC").Find(&assinaturas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar assinaturas do contato"})
		return
	}

	c.JSON(http.StatusOK, assinaturas)
}

// UpdateAssinaturaStatus atualiza apenas o status da assinatura
func (h *AssinaturasHandler) UpdateAssinaturaStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	assinaturaID := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var assinatura models.Assinatura
	if err := h.db.Where("id = ? AND usuario_id = ?", assinaturaID, userID).First(&assinatura).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assinatura não encontrada"})
		return
	}

	if err := h.db.Model(&assinatura).Updates(map[string]interface{}{
		"status": req.Status,
		"atualizado_em": time.Now(),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status atualizado com sucesso"})
}

// GetAssinaturasVencendoSoon busca assinaturas que vencem em breve
func (h *AssinaturasHandler) GetAssinaturasVencendoSoon(c *gin.Context) {
	userID := c.GetString("user_id")

	var assinaturas []models.Assinatura
	if err := h.db.Where("usuario_id = ? AND data_fim <= ?", userID, time.Now().AddDate(0, 0, 7)).Order("data_fim ASC").Find(&assinaturas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar assinaturas vencendo"})
		return
	}

	c.JSON(http.StatusOK, assinaturas)
}
