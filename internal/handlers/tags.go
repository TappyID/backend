package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tappyone/internal/models"
	"tappyone/internal/services"
)

type TagsHandler struct {
	db      *gorm.DB
	authSvc *services.AuthService
}

func NewTagsHandler(db *gorm.DB, authSvc *services.AuthService) *TagsHandler {
	return &TagsHandler{
		db:      db,
		authSvc: authSvc,
	}
}

// ListarTags - GET /api/tags
func (h *TagsHandler) ListarTags(c *gin.Context) {
	var tags []models.Tag

	// Buscar todas as tags ordenadas por nome
	if err := h.db.Order("nome ASC").Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Erro ao buscar tags",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tags,
	})
}

// CriarTag - POST /api/tags
func (h *TagsHandler) CriarTag(c *gin.Context) {
	var tag models.Tag

	if err := c.ShouldBindJSON(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Dados inválidos",
			"details": err.Error(),
		})
		return
	}

	// Validar se nome não está vazio
	if tag.Nome == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Nome da tag é obrigatório",
		})
		return
	}

	// Verificar se tag já existe
	var existeTag models.Tag
	if err := h.db.Where("nome = ?", tag.Nome).First(&existeTag).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Tag com este nome já existe",
		})
		return
	}

	// Criar tag
	if err := h.db.Create(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Erro ao criar tag",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    tag,
		"message": "Tag criada com sucesso",
	})
}

// AtualizarTag - PUT /api/tags/:id
func (h *TagsHandler) AtualizarTag(c *gin.Context) {
	id := c.Param("id")
	var tag models.Tag

	// Buscar tag existente
	if err := h.db.First(&tag, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Tag não encontrada",
		})
		return
	}

	var dadosAtualizacao models.Tag
	if err := c.ShouldBindJSON(&dadosAtualizacao); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Dados inválidos",
			"details": err.Error(),
		})
		return
	}

	// Verificar se novo nome já existe (apenas se nome foi alterado)
	if dadosAtualizacao.Nome != "" && dadosAtualizacao.Nome != tag.Nome {
		var existeTag models.Tag
		if err := h.db.Where("nome = ? AND id != ?", dadosAtualizacao.Nome, id).First(&existeTag).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "Tag com este nome já existe",
			})
			return
		}
		tag.Nome = dadosAtualizacao.Nome
	}

	if dadosAtualizacao.Descricao != "" {
		tag.Descricao = dadosAtualizacao.Descricao
	}

	if dadosAtualizacao.Cor != "" {
		tag.Cor = dadosAtualizacao.Cor
	}

	tag.Favorito = dadosAtualizacao.Favorito

	// Salvar alterações
	if err := h.db.Save(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Erro ao atualizar tag",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tag,
		"message": "Tag atualizada com sucesso",
	})
}

// DeletarTag - DELETE /api/tags/:id
func (h *TagsHandler) DeletarTag(c *gin.Context) {
	id := c.Param("id")
	var tag models.Tag

	// Buscar tag existente
	if err := h.db.First(&tag, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Tag não encontrada",
		})
		return
	}

	// Verificar se tag está sendo usada
	var contatoTagCount int64
	h.db.Model(&models.ContatoTag{}).Where("tag_id = ?", id).Count(&contatoTagCount)
	
	var quadroTagCount int64
	h.db.Model(&models.QuadroTag{}).Where("tag_id = ?", id).Count(&quadroTagCount)

	if contatoTagCount > 0 || quadroTagCount > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Não é possível deletar tag que está sendo usada",
			"details": "Tag está associada a contatos ou quadros",
		})
		return
	}

	// Deletar tag
	if err := h.db.Delete(&tag).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Erro ao deletar tag",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tag deletada com sucesso",
	})
}

// ObterTag - GET /api/tags/:id
func (h *TagsHandler) ObterTag(c *gin.Context) {
	id := c.Param("id")
	var tag models.Tag

	if err := h.db.First(&tag, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Tag não encontrada",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tag,
	})
}

// ListarTagsFavoritas - GET /api/tags/favoritas
func (h *TagsHandler) ListarTagsFavoritas(c *gin.Context) {
	var tags []models.Tag

	if err := h.db.Where("favorito = ?", true).Order("nome ASC").Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Erro ao buscar tags favoritas",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tags,
	})
}

// ListarTagsPopulares - GET /api/tags/populares
func (h *TagsHandler) ListarTagsPopulares(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	type TagComUso struct {
		models.Tag
		TotalUsos int64 `json:"totalUsos"`
	}

	var tagsComUso []TagComUso

	// Query para buscar tags ordenadas por uso
	if err := h.db.Raw(`
		SELECT t.*, 
		       (COALESCE(ct.contato_count, 0) + COALESCE(qt.quadro_count, 0)) as total_usos
		FROM tags t
		LEFT JOIN (
			SELECT tag_id, COUNT(*) as contato_count 
			FROM contato_tags 
			GROUP BY tag_id
		) ct ON t.id = ct.tag_id
		LEFT JOIN (
			SELECT tag_id, COUNT(*) as quadro_count 
			FROM quadro_tags 
			GROUP BY tag_id
		) qt ON t.id = qt.tag_id
		ORDER BY total_usos DESC, t.nome ASC
		LIMIT ?
	`, limit).Scan(&tagsComUso).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Erro ao buscar tags populares",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tagsComUso,
	})
}
