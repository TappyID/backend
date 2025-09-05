package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tappyone/internal/models"
)

type SessoesWhatsAppHandler struct {
	db *gorm.DB
}

func NewSessoesWhatsAppHandler(db *gorm.DB) *SessoesWhatsAppHandler {
	return &SessoesWhatsAppHandler{db: db}
}

// ListSessoesWhatsApp lista todas as sessões WhatsApp do usuário
func (h *SessoesWhatsAppHandler) ListSessoesWhatsApp(c *gin.Context) {
	userID := c.GetString("user_id")

	var sessoes []models.SessaoWhatsApp
	if err := h.db.Where("usuario_id = ?", userID).Find(&sessoes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar sessões WhatsApp"})
		return
	}

	c.JSON(http.StatusOK, sessoes)
}

// GetSessaoWhatsApp obtém uma sessão WhatsApp específica
func (h *SessoesWhatsAppHandler) GetSessaoWhatsApp(c *gin.Context) {
	userID := c.GetString("user_id")
	sessaoID := c.Param("id")

	var sessao models.SessaoWhatsApp
	if err := h.db.Where("id = ? AND usuario_id = ?", sessaoID, userID).First(&sessao).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sessão WhatsApp não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar sessão WhatsApp"})
		return
	}

	c.JSON(http.StatusOK, sessao)
}

// CreateSessaoWhatsApp cria uma nova sessão WhatsApp
func (h *SessoesWhatsAppHandler) CreateSessaoWhatsApp(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		NomeSessao string `json:"nomeSessao" binding:"required"`
		Status     string `json:"status"`
		Ativo      bool   `json:"ativo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Se esta sessão vai ser ativa, desativar outras
	if req.Ativo {
		h.db.Model(&models.SessaoWhatsApp{}).Where("usuario_id = ?", userID).Update("ativo", false)
	}

	sessao := models.SessaoWhatsApp{
		UsuarioID:  userID,
		NomeSessao: req.NomeSessao,
		Status:     models.StatusSessao(req.Status),
		Ativo:      req.Ativo,
	}

	if err := h.db.Create(&sessao).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar sessão WhatsApp"})
		return
	}

	c.JSON(http.StatusCreated, sessao)
}

// UpdateSessaoWhatsApp atualiza uma sessão WhatsApp
func (h *SessoesWhatsAppHandler) UpdateSessaoWhatsApp(c *gin.Context) {
	userID := c.GetString("user_id")
	sessaoID := c.Param("id")

	var req struct {
		NomeSessao *string `json:"nomeSessao"`
		Status     *string `json:"status"`
		Ativo      *bool   `json:"ativo"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar se a sessão existe e pertence ao usuário
	var sessao models.SessaoWhatsApp
	if err := h.db.Where("id = ? AND usuario_id = ?", sessaoID, userID).First(&sessao).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sessão WhatsApp não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar sessão WhatsApp"})
		return
	}

	// Se esta sessão vai ser ativa, desativar outras
	if req.Ativo != nil && *req.Ativo {
		h.db.Model(&models.SessaoWhatsApp{}).Where("usuario_id = ? AND id != ?", userID, sessaoID).Update("ativo", false)
	}

	// Preparar dados para atualização
	updates := make(map[string]interface{})
	if req.NomeSessao != nil {
		updates["nome_sessao"] = *req.NomeSessao
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Ativo != nil {
		updates["ativo"] = *req.Ativo
	}

	if err := h.db.Model(&sessao).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar sessão WhatsApp"})
		return
	}

	// Buscar sessão atualizada
	if err := h.db.Where("id = ?", sessaoID).First(&sessao).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar sessão atualizada"})
		return
	}

	c.JSON(http.StatusOK, sessao)
}

// DeleteSessaoWhatsApp remove uma sessão WhatsApp
func (h *SessoesWhatsAppHandler) DeleteSessaoWhatsApp(c *gin.Context) {
	userID := c.GetString("user_id")
	sessaoID := c.Param("id")

	// Verificar se a sessão existe e pertence ao usuário
	var sessao models.SessaoWhatsApp
	if err := h.db.Where("id = ? AND usuario_id = ?", sessaoID, userID).First(&sessao).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sessão WhatsApp não encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar sessão WhatsApp"})
		return
	}

	// Verificar se existem contatos usando esta sessão
	var contatosCount int64
	if err := h.db.Model(&models.Contato{}).Where("sessao_whatsapp_id = ?", sessaoID).Count(&contatosCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao verificar contatos"})
		return
	}

	if contatosCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Não é possível excluir sessão com contatos associados"})
		return
	}

	if err := h.db.Delete(&sessao).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao deletar sessão WhatsApp"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sessão WhatsApp deletada com sucesso"})
}
