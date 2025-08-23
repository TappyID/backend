package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tappyone/internal/models"
)

// OrcamentosHandler gerencia os orçamentos
type OrcamentosHandler struct {
	db *gorm.DB
}

func NewOrcamentosHandler(db *gorm.DB) *OrcamentosHandler {
	return &OrcamentosHandler{db: db}
}

// ListOrcamentos lista todos os orçamentos do usuário
func (h *OrcamentosHandler) ListOrcamentos(c *gin.Context) {
	userID := c.GetString("user_id")
	contatoJID := c.Query("contato_id")
	status := c.Query("status")

	var orcamentos []models.Orcamento
	query := h.db.Where("usuario_id = ?", userID).Preload("Itens")
	
	if contatoJID != "" {
		// Converter JID para número de telefone e buscar contato
		numeroTelefone := strings.Replace(contatoJID, "@c.us", "", 1)
		
		// Buscar contato pelo número de telefone
		var contato models.Contato
		if err := h.db.Where("numero_telefone = ?", numeroTelefone).First(&contato).Error; err == nil {
			// Se contato encontrado, filtrar por UUID do contato
			query = query.Where("contato_id = ?", contato.ID)
		} else {
			// Se contato não encontrado, não retornar nenhum orçamento
			c.JSON(http.StatusOK, []models.Orcamento{})
			return
		}
	}
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if err := query.Order("data DESC").Find(&orcamentos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar orçamentos"})
		return
	}

	c.JSON(http.StatusOK, orcamentos)
}

// GetOrcamento busca um orçamento específico
func (h *OrcamentosHandler) GetOrcamento(c *gin.Context) {
	userID := c.GetString("user_id")
	orcamentoID := c.Param("id")

	var orcamento models.Orcamento
	if err := h.db.Where("id = ? AND usuario_id = ?", orcamentoID, userID).
		Preload("Itens").First(&orcamento).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Orçamento não encontrado"})
		return
	}

	c.JSON(http.StatusOK, orcamento)
}

// CreateOrcamento cria um novo orçamento
func (h *OrcamentosHandler) CreateOrcamento(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		Titulo     string                  `json:"titulo" binding:"required"`
		Data       time.Time               `json:"data" binding:"required"`
		Tipo       string                  `json:"tipo" binding:"required"`
		Observacao *string                 `json:"observacao"`
		ContatoID  string                  `json:"contato_id" binding:"required"`
		Itens      []models.OrcamentoItem  `json:"itens"`
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

	// Calcular valor total
	var valorTotal float64
	for _, item := range req.Itens {
		valorTotal += item.Subtotal
	}

	orcamento := models.Orcamento{
		Titulo:     req.Titulo,
		Data:       req.Data,
		Tipo:       models.TipoOrcamento(req.Tipo),
		Observacao: req.Observacao,
		ValorTotal: valorTotal,
		Status:     "PENDENTE",
		UsuarioID:  userID,
		ContatoID:  contato.ID, // Usar ID do contato encontrado/criado
		Itens:      req.Itens,
	}

	if err := h.db.Create(&orcamento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar orçamento"})
		return
	}

	c.JSON(http.StatusCreated, orcamento)
}

// UpdateOrcamento atualiza um orçamento existente
func (h *OrcamentosHandler) UpdateOrcamento(c *gin.Context) {
	userID := c.GetString("user_id")
	orcamentoID := c.Param("id")

	var req struct {
		Titulo     *string                 `json:"titulo"`
		Data       *time.Time              `json:"data"`
		Tipo       *string                 `json:"tipo"`
		Observacao *string                 `json:"observacao"`
		Status     *string                 `json:"status"`
		Itens      *[]models.OrcamentoItem `json:"itens"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var orcamento models.Orcamento
	if err := h.db.Where("id = ? AND usuario_id = ?", orcamentoID, userID).
		Preload("Itens").First(&orcamento).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Orçamento não encontrado"})
		return
	}

	// Atualizar campos fornecidos
	updates := make(map[string]interface{})
	if req.Titulo != nil {
		updates["titulo"] = *req.Titulo
	}
	if req.Data != nil {
		updates["data"] = *req.Data
	}
	if req.Tipo != nil {
		updates["tipo"] = *req.Tipo
	}
	if req.Observacao != nil {
		updates["observacao"] = *req.Observacao
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	updates["atualizado_em"] = time.Now()

	// Atualizar itens se fornecidos
	if req.Itens != nil {
		// Excluir itens existentes
		if err := h.db.Where("orcamento_id = ?", orcamentoID).Delete(&models.OrcamentoItem{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar itens"})
			return
		}

		// Criar novos itens
		var valorTotal float64
		for _, item := range *req.Itens {
			item.OrcamentoID = orcamentoID
			if err := h.db.Create(&item).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar item"})
				return
			}
			valorTotal += item.Subtotal
		}
		updates["valor_total"] = valorTotal
	}

	if err := h.db.Model(&orcamento).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar orçamento"})
		return
	}

	// Recarregar dados atualizados
	if err := h.db.Where("id = ?", orcamentoID).Preload("Itens").First(&orcamento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao recarregar orçamento"})
		return
	}

	c.JSON(http.StatusOK, orcamento)
}

// DeleteOrcamento exclui um orçamento
func (h *OrcamentosHandler) DeleteOrcamento(c *gin.Context) {
	userID := c.GetString("user_id")
	orcamentoID := c.Param("id")

	var orcamento models.Orcamento
	if err := h.db.Where("id = ? AND usuario_id = ?", orcamentoID, userID).First(&orcamento).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Orçamento não encontrado"})
		return
	}

	// Excluir itens relacionados primeiro (cascade delete)
	if err := h.db.Where("orcamento_id = ?", orcamentoID).Delete(&models.OrcamentoItem{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir itens do orçamento"})
		return
	}

	// Excluir orçamento
	if err := h.db.Delete(&orcamento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir orçamento"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Orçamento excluído com sucesso"})
}

// GetOrcamentosByContato busca todos os orçamentos de um contato específico
func (h *OrcamentosHandler) GetOrcamentosByContato(c *gin.Context) {
	userID := c.GetString("user_id")
	contatoID := c.Param("contato_id")

	var orcamentos []models.Orcamento
	if err := h.db.Where("usuario_id = ? AND contato_id = ?", userID, contatoID).
		Preload("Itens").Order("data DESC").Find(&orcamentos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar orçamentos do contato"})
		return
	}

	c.JSON(http.StatusOK, orcamentos)
}

// UpdateOrcamentoStatus atualiza apenas o status do orçamento
func (h *OrcamentosHandler) UpdateOrcamentoStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	orcamentoID := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var orcamento models.Orcamento
	if err := h.db.Where("id = ? AND usuario_id = ?", orcamentoID, userID).First(&orcamento).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Orçamento não encontrado"})
		return
	}

	if err := h.db.Model(&orcamento).Updates(map[string]interface{}{
		"status": req.Status,
		"atualizado_em": time.Now(),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status atualizado com sucesso"})
}
