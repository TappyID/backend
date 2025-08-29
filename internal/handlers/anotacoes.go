package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"tappyone/internal/models"
)

// AnotacoesHandler gerencia as anotações
type AnotacoesHandler struct {
	db *gorm.DB
}

func NewAnotacoesHandler(db *gorm.DB) *AnotacoesHandler {
	return &AnotacoesHandler{db: db}
}

// ListAnotacoes lista todas as anotações do usuário
func (h *AnotacoesHandler) ListAnotacoes(c *gin.Context) {
	// Verificar autenticação
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	userIDStr := userID.(string)
	contatoJID := c.Query("contato_id")

	var anotacoes []models.Anotacao
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
			// Se contato não encontrado, não retornar nenhuma anotação
			c.JSON(http.StatusOK, []models.Anotacao{})
			return
		}
	}
	
	if err := query.Order("criado_em DESC").Find(&anotacoes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar anotações"})
		return
	}

	c.JSON(http.StatusOK, anotacoes)
}

// GetAnotacao busca uma anotação específica
func (h *AnotacoesHandler) GetAnotacao(c *gin.Context) {
	userID := c.GetString("user_id")
	anotacaoID := c.Param("id")

	var anotacao models.Anotacao
	if err := h.db.Where("id = ? AND usuario_id = ?", anotacaoID, userID).First(&anotacao).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Anotação não encontrada"})
		return
	}

	c.JSON(http.StatusOK, anotacao)
}

// CreateAnotacao cria uma nova anotação
func (h *AnotacoesHandler) CreateAnotacao(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		Titulo    string `json:"titulo" binding:"required"`
		Conteudo  string `json:"conteudo" binding:"required"`
		Importante bool  `json:"importante"`
		ContatoID string `json:"contato_id" binding:"required"`
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

	anotacao := models.Anotacao{
		Titulo:     req.Titulo,
		Conteudo:   req.Conteudo,
		Importante: req.Importante,
		UsuarioID:  userID,
		ContatoID:  contato.ID, // Usar ID do contato encontrado/criado
	}

	if err := h.db.Create(&anotacao).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar anotação"})
		return
	}

	c.JSON(http.StatusCreated, anotacao)
}

// UpdateAnotacao atualiza uma anotação existente
func (h *AnotacoesHandler) UpdateAnotacao(c *gin.Context) {
	userID := c.GetString("user_id")
	anotacaoID := c.Param("id")

	var req struct {
		Titulo     *string `json:"titulo"`
		Conteudo   *string `json:"conteudo"`
		Importante *bool   `json:"importante"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var anotacao models.Anotacao
	if err := h.db.Where("id = ? AND usuario_id = ?", anotacaoID, userID).First(&anotacao).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Anotação não encontrada"})
		return
	}

	// Atualizar campos fornecidos
	updates := make(map[string]interface{})
	if req.Titulo != nil {
		updates["titulo"] = *req.Titulo
	}
	if req.Conteudo != nil {
		updates["conteudo"] = *req.Conteudo
	}
	if req.Importante != nil {
		updates["importante"] = *req.Importante
	}
	updates["atualizado_em"] = time.Now()

	if err := h.db.Model(&anotacao).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar anotação"})
		return
	}

	// Recarregar dados atualizados
	if err := h.db.Where("id = ?", anotacaoID).First(&anotacao).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao recarregar anotação"})
		return
	}

	c.JSON(http.StatusOK, anotacao)
}

// DeleteAnotacao exclui uma anotação
func (h *AnotacoesHandler) DeleteAnotacao(c *gin.Context) {
	userID := c.GetString("user_id")
	anotacaoID := c.Param("id")

	var anotacao models.Anotacao
	if err := h.db.Where("id = ? AND usuario_id = ?", anotacaoID, userID).First(&anotacao).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Anotação não encontrada"})
		return
	}

	if err := h.db.Delete(&anotacao).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir anotação"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Anotação excluída com sucesso"})
}

// GetAnotacoesByContato busca todas as anotações de um contato específico
func (h *AnotacoesHandler) GetAnotacoesByContato(c *gin.Context) {
	userID := c.GetString("user_id")
	contatoID := c.Param("contato_id")

	var anotacoes []models.Anotacao
	if err := h.db.Where("usuario_id = ? AND contato_id = ?", userID, contatoID).
		Order("importante DESC, criado_em DESC").Find(&anotacoes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar anotações do contato"})
		return
	}

	c.JSON(http.StatusOK, anotacoes)
}
