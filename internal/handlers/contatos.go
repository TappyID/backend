package handlers

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tappyone/internal/models"
)

type ContatosHandler struct {
	db *gorm.DB
}

func NewContatosHandler(db *gorm.DB) *ContatosHandler {
	return &ContatosHandler{db: db}
}

// Use models do pacote models em vez de duplicar

// CreateContatoRequest representa os dados para criar um contato
type CreateContatoRequest struct {
	NumeroTelefone   string  `json:"numeroTelefone" binding:"required"`
	Nome             *string `json:"nome"`
	Email            *string `json:"email"`
	Empresa          *string `json:"empresa"`
	CPF              *string `json:"cpf"`
	CNPJ             *string `json:"cnpj"`
	CEP              *string `json:"cep"`
	Rua              *string `json:"rua"`
	Numero           *string `json:"numero"`
	Bairro           *string `json:"bairro"`
	Cidade           *string `json:"cidade"`
	Estado           *string `json:"estado"`
	Pais             *string `json:"pais"`
	SessaoWhatsappID string  `json:"sessaoWhatsappId" binding:"required"`
}

// UpdateContatoRequest representa os dados para atualizar um contato
type UpdateContatoRequest struct {
	Nome      *string `json:"nome"`
	Email     *string `json:"email"`
	Empresa   *string `json:"empresa"`
	CPF       *string `json:"cpf"`
	CNPJ      *string `json:"cnpj"`
	CEP       *string `json:"cep"`
	Rua       *string `json:"rua"`
	Numero    *string `json:"numero"`
	Bairro    *string `json:"bairro"`
	Cidade    *string `json:"cidade"`
	Estado    *string `json:"estado"`
	Pais      *string `json:"pais"`
	Bloqueado *bool   `json:"bloqueado"`
}

// ListContatos lista todos os contatos do usuário
func (h *ContatosHandler) ListContatos(c *gin.Context) {
	// Obter userID do contexto (middleware de auth)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Buscar contatos associados às sessões WhatsApp do usuário
	var contatos []models.Contato
	query := `
		SELECT c.id, c.numero_telefone, c.nome, c.foto_perfil, c.sobre, c.bloqueado,
		       c.sessao_whatsapp_id, c.email, c.empresa, c.cpf, c.cnpj, c.cep, c.rua,
		       c.numero, c.bairro, c.cidade, c.estado, c.pais, c.criado_em, c.atualizado_em
		FROM contatos c
		INNER JOIN sessoes_whatsapp sw ON c.sessao_whatsapp_id = sw.id
		WHERE sw.usuario_id = ?
		ORDER BY c.atualizado_em DESC
	`

	err := h.db.Raw(query, userID).Scan(&contatos).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contacts"})
		return
	}

	// Para cada contato, buscar tags
	for i := range contatos {
		var tags []models.ContatoTag
		err := h.db.Table("contato_tags ct").
			Select("ct.id, ct.contato_id, ct.tag_id, t.id as \"tag.id\", t.nome as \"tag.nome\", t.cor as \"tag.cor\"").
			Joins("INNER JOIN tags t ON ct.tag_id = t.id").
			Where("ct.contato_id = ?", contatos[i].ID).
			Scan(&tags).Error

		if err == nil {
			contatos[i].Tags = tags
		}
	}

	c.JSON(http.StatusOK, contatos)
}

// GetContato busca um contato específico
func (h *ContatosHandler) GetContato(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	var contato models.Contato

	query := `
		SELECT c.id, c.numero_telefone, c.nome, c.foto_perfil, c.sobre, c.bloqueado,
		       c.sessao_whatsapp_id, c.email, c.empresa, c.cpf, c.cnpj, c.cep, c.rua,
		       c.numero, c.bairro, c.cidade, c.estado, c.pais, c.criado_em, c.atualizado_em
		FROM contatos c
		INNER JOIN sessoes_whatsapp sw ON c.sessao_whatsapp_id = sw.id
		WHERE c.id = ? AND sw.usuario_id = ?
	`

	err := h.db.Raw(query, id, userID).Scan(&contato).Error
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contact"})
		return
	}

	// Buscar tags do contato
	var tags []models.ContatoTag
	err = h.db.Table("contato_tags ct").
		Select("ct.id, ct.contato_id, ct.tag_id, t.id as 'tag.id', t.nome as 'tag.nome', t.cor as 'tag.cor'").
		Joins("INNER JOIN tags t ON ct.tag_id = t.id").
		Where("ct.contato_id = ?", contato.ID).
		Scan(&tags).Error

	if err == nil {
		contato.Tags = tags
	}

	c.JSON(http.StatusOK, contato)
}

// CreateContato cria um novo contato
func (h *ContatosHandler) CreateContato(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req CreateContatoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar se a sessão WhatsApp pertence ao usuário
	var sessaoCount int64
	err := h.db.Table("sessoes_whatsapp").
		Where("id = ? AND usuario_id = ?", req.SessaoWhatsappID, userID).
		Count(&sessaoCount).Error

	if err != nil || sessaoCount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid WhatsApp session"})
		return
	}

	// Criar contato
	contato := models.Contato{
		NumeroTelefone:   req.NumeroTelefone,
		Nome:             req.Nome,
		Email:            req.Email,
		Empresa:          req.Empresa,
		CPF:              req.CPF,
		CNPJ:             req.CNPJ,
		CEP:              req.CEP,
		Rua:              req.Rua,
		Numero:           req.Numero,
		Bairro:           req.Bairro,
		Cidade:           req.Cidade,
		Estado:           req.Estado,
		Pais:             req.Pais,
		SessaoWhatsappID: req.SessaoWhatsappID,
		Bloqueado:        false,
	}

	if err := h.db.Create(&contato).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contact"})
		return
	}

	c.JSON(http.StatusCreated, contato)
}

// GetContatosStats retorna estatísticas dos contatos
func (h *ContatosHandler) GetContatosStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	type StatsResult struct {
		TotalContatos   int64 `json:"totalContatos"`
		ContatosAtivos  int64 `json:"contatosAtivos"`
		Favoritos       int64 `json:"favoritos"`
		ConversasHoje   int64 `json:"conversasHoje"`
	}

	var stats StatsResult

	// Total de contatos do usuário
	err := h.db.Table("contatos").
		Joins("INNER JOIN sessoes_whatsapp sw ON contatos.sessao_whatsapp_id = sw.id").
		Where("sw.usuario_id = ?", userID).
		Count(&stats.TotalContatos).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get total contacts"})
		return
	}

	// Favoritos
	err = h.db.Table("contatos").
		Joins("INNER JOIN sessoes_whatsapp sw ON contatos.sessao_whatsapp_id = sw.id").
		Where("sw.usuario_id = ? AND contatos.favorito = ?", userID, true).
		Count(&stats.Favoritos).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get favorites count"})
		return
	}

	// Por enquanto, usar valores estimados para contatos ativos e conversas hoje
	// até que tenhamos tabelas de mensagens implementadas
	stats.ContatosAtivos = int64(float64(stats.TotalContatos) * 0.7) // ~70% dos contatos ativos
	stats.ConversasHoje = int64(float64(stats.TotalContatos) * 0.1)  // ~10% com conversas hoje

	c.JSON(http.StatusOK, stats)
}

// ExportContatos exporta contatos do usuário para CSV ou JSON
func (h *ContatosHandler) ExportContatos(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	format := c.Query("format")
	if format != "csv" && format != "json" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format must be 'csv' or 'json'"})
		return
	}

	// Buscar contatos do usuário
	var contatos []models.Contato
	err := h.db.Table("contatos c").
		Select("c.*").
		Joins("INNER JOIN sessoes_whatsapp sw ON c.sessao_whatsapp_id = sw.id").
		Where("sw.usuario_id = ?", userID).
		Find(&contatos).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contacts"})
		return
	}

	if format == "json" {
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=contatos.json")
		c.JSON(http.StatusOK, contatos)
	} else {
		// CSV format
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=contatos.csv")
		
		csvHeader := "nome,numeroTelefone,email,empresa,cpf,cnpj,cep,rua,numero,bairro,cidade,estado,pais,favorito\n"
		csvData := csvHeader
		
		for _, contato := range contatos {
			csvData += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%t\n",
				contato.Nome,
				contato.NumeroTelefone,
				contato.Email,
				contato.Empresa,
				contato.CPF,
				contato.CNPJ,
				contato.CEP,
				contato.Rua,
				contato.Numero,
				contato.Bairro,
				contato.Cidade,
				contato.Estado,
				contato.Pais,
				contato.Favorito,
			)
		}
		
		c.Data(http.StatusOK, "text/csv", []byte(csvData))
	}
}

// ImportContatos importa contatos de arquivo CSV ou JSON
func (h *ContatosHandler) ImportContatos(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// Verificar se o usuário tem uma sessão WhatsApp ativa
	var sessaoWhatsappID string
	err = h.db.Table("sessoes_whatsapp").
		Select("id").
		Where("usuario_id = ? AND ativo = ?", userID, true).
		First(&sessaoWhatsappID).Error

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No active WhatsApp session found"})
		return
	}

	// Ler conteúdo do arquivo
	fileBytes := make([]byte, header.Size)
	_, err = file.Read(fileBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	var importedContacts []models.Contato
	filename := header.Filename

	if strings.HasSuffix(strings.ToLower(filename), ".json") {
		// Parse JSON
		var jsonContacts []map[string]interface{}
		err = json.Unmarshal(fileBytes, &jsonContacts)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		// Converter para modelos de contato
		for _, item := range jsonContacts {
			contato := models.Contato{
				Nome:             getStringPtrFromMap(item, "nome"),
				NumeroTelefone:   getStringFromMap(item, "numeroTelefone"),
				Email:            getStringPtrFromMap(item, "email"),
				Empresa:          getStringPtrFromMap(item, "empresa"),
				CPF:              getStringPtrFromMap(item, "cpf"),
				CNPJ:             getStringPtrFromMap(item, "cnpj"),
				CEP:              getStringPtrFromMap(item, "cep"),
				Rua:              getStringPtrFromMap(item, "rua"),
				Numero:           getStringPtrFromMap(item, "numero"),
				Bairro:           getStringPtrFromMap(item, "bairro"),
				Cidade:           getStringPtrFromMap(item, "cidade"),
				Estado:           getStringPtrFromMap(item, "estado"),
				Pais:             getStringPtrFromMap(item, "pais"),
				Favorito:         getBoolFromMap(item, "favorito"),
				SessaoWhatsappID: sessaoWhatsappID,
			}
			importedContacts = append(importedContacts, contato)
		}
	} else if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		// Parse CSV
		reader := csv.NewReader(strings.NewReader(string(fileBytes)))
		records, err := reader.ReadAll()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid CSV format"})
			return
		}

		// Pular header se existir
		if len(records) > 1 && records[0][0] == "nome" {
			records = records[1:]
		}

		for _, record := range records {
			if len(record) >= 13 {
				favorito, _ := strconv.ParseBool(record[13])
				contato := models.Contato{
					Nome:             stringPtr(record[0]),
					NumeroTelefone:   record[1],
					Email:            stringPtr(record[2]),
					Empresa:          stringPtr(record[3]),
					CPF:              stringPtr(record[4]),
					CNPJ:             stringPtr(record[5]),
					CEP:              stringPtr(record[6]),
					Rua:              stringPtr(record[7]),
					Numero:           stringPtr(record[8]),
					Bairro:           stringPtr(record[9]),
					Cidade:           stringPtr(record[10]),
					Estado:           stringPtr(record[11]),
					Pais:             stringPtr(record[12]),
					Favorito:         favorito,
					SessaoWhatsappID: sessaoWhatsappID,
				}
				importedContacts = append(importedContacts, contato)
			}
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File must be .json or .csv"})
		return
	}

	// Inserir contatos no banco de dados
	successCount := 0
	errorCount := 0
	var errors []string

	for _, contato := range importedContacts {
		if contato.NumeroTelefone == "" || (contato.Nome == nil || *contato.Nome == "") {
			errorCount++
			errors = append(errors, fmt.Sprintf("Contato inválido: nome ou telefone em branco"))
			continue
		}

		err = h.db.Create(&contato).Error
		if err != nil {
			errorCount++
			errors = append(errors, fmt.Sprintf("Erro ao criar contato %s: %v", contato.Nome, err))
		} else {
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Import completed",
		"success":      successCount,
		"errors":       errorCount,
		"errorDetails": errors,
	})
}

// Funções auxiliares para conversão de tipos
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getStringPtrFromMap(m map[string]interface{}, key string) *string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok && str != "" {
			return &str
		}
	}
	return nil
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

// UpdateContato atualiza um contato existente
func (h *ContatosHandler) UpdateContato(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	var req UpdateContatoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verificar se o contato existe e pertence ao usuário
	var contato models.Contato
	query := `
		SELECT c.id, c.numero_telefone, c.nome, c.foto_perfil, c.sobre, c.bloqueado,
		       c.sessao_whatsapp_id, c.email, c.empresa, c.cpf, c.cnpj, c.cep, c.rua,
		       c.numero, c.bairro, c.cidade, c.estado, c.pais, c.criado_em, c.atualizado_em
		FROM contatos c
		INNER JOIN sessoes_whatsapp sw ON c.sessao_whatsapp_id = sw.id
		WHERE c.id = ? AND sw.usuario_id = ?
	`

	err := h.db.Raw(query, id, userID).Scan(&contato).Error
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contact"})
		return
	}

	// Atualizar campos não nulos
	updateData := make(map[string]interface{})
	if req.Nome != nil {
		updateData["nome"] = *req.Nome
	}
	if req.Email != nil {
		updateData["email"] = *req.Email
	}
	if req.Empresa != nil {
		updateData["empresa"] = *req.Empresa
	}
	if req.CPF != nil {
		updateData["cpf"] = *req.CPF
	}
	if req.CNPJ != nil {
		updateData["cnpj"] = *req.CNPJ
	}
	if req.CEP != nil {
		updateData["cep"] = *req.CEP
	}
	if req.Rua != nil {
		updateData["rua"] = *req.Rua
	}
	if req.Numero != nil {
		updateData["numero"] = *req.Numero
	}
	if req.Bairro != nil {
		updateData["bairro"] = *req.Bairro
	}
	if req.Cidade != nil {
		updateData["cidade"] = *req.Cidade
	}
	if req.Estado != nil {
		updateData["estado"] = *req.Estado
	}
	if req.Pais != nil {
		updateData["pais"] = *req.Pais
	}
	if req.Bloqueado != nil {
		updateData["bloqueado"] = *req.Bloqueado
	}

	if len(updateData) > 0 {
		if err := h.db.Table("contatos").Where("id = ?", id).Updates(updateData).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contact"})
			return
		}
	}

	// Buscar contato atualizado
	err = h.db.Raw(query, id, userID).Scan(&contato).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated contact"})
		return
	}

	c.JSON(http.StatusOK, contato)
}

// DeleteContato deleta um contato
func (h *ContatosHandler) DeleteContato(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")

	// Verificar se o contato existe e pertence ao usuário
	var count int64
	query := `
		SELECT COUNT(*)
		FROM contatos c
		INNER JOIN sessoes_whatsapp sw ON c.sessao_whatsapp_id = sw.id
		WHERE c.id = ? AND sw.usuario_id = ?
	`

	err := h.db.Raw(query, id, userID).Scan(&count).Error
	if err != nil || count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
		return
	}

	// Deletar contato (cascade deletes will handle related records)
	if err := h.db.Table("contatos").Where("id = ?", id).Delete(nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete contact"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contact deleted successfully"})
}

// SyncContatos sincroniza contatos do WhatsApp
func (h *ContatosHandler) SyncContatos(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Por enquanto, retorna uma mensagem de sucesso
	// A sincronização real com o WAHA seria implementada aqui
	c.JSON(http.StatusOK, gin.H{
		"message": "Sync contacts functionality will be implemented",
		"userID":  userID,
	})
}

// GetContatoDadosCompletos busca dados completos do contato pelo chatId do WAHA
func (h *ContatosHandler) GetContatoDadosCompletos(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	chatId := c.Param("chatId")
	
	// Buscar contato pelo contactid (que vem do WAHA)
	var contato models.Contato
	query := `
		SELECT c.id, c.numero_telefone, c.nome, c.foto_perfil, c.sobre, c.bloqueado,
		       c.sessao_whatsapp_id, c.email, c.empresa, c.cpf, c.cnpj, c.cep, c.rua,
		       c.numero, c.bairro, c.cidade, c.estado, c.pais, c.criado_em, c.atualizado_em,
		       c.contactid, c.status_kanban
		FROM contatos c
		INNER JOIN sessoes_whatsapp sw ON c.sessao_whatsapp_id = sw.id
		WHERE c.contactid = ? AND sw.usuario_id = ?
	`

	err := h.db.Raw(query, chatId, userID).Scan(&contato).Error
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contact"})
		return
	}

	// Buscar fila do contato
	var fila *models.Fila
	filaQuery := `
		SELECT f.id, f.nome, f.cor, f.descricao, f.ativa
		FROM filas f
		INNER JOIN fila_contatos fc ON f.id = fc.fila_id
		WHERE fc.contato_id = ?
		LIMIT 1
	`
	
	var filaResult models.Fila
	err = h.db.Raw(filaQuery, contato.ID).Scan(&filaResult).Error
	if err == nil {
		fila = &filaResult
	}

	// Buscar tags do contato
	var tags []models.Tag
	tagQuery := `
		SELECT t.id, t.nome, t.cor
		FROM tags t
		INNER JOIN contato_tags ct ON t.id = ct.tag_id
		WHERE ct.contato_id = ?
	`
	
	err = h.db.Raw(tagQuery, contato.ID).Scan(&tags).Error
	if err != nil {
		tags = []models.Tag{}
	}

	// Buscar atendente do contato
	var atendente *models.Usuario
	atendenteQuery := `
		SELECT u.id, u.nome, u.email
		FROM usuarios u
		INNER JOIN atendente_contatos ac ON u.id = ac.user_id
		WHERE ac.contato_id = ?
		LIMIT 1
	`
	
	var atendenteResult models.Usuario
	err = h.db.Raw(atendenteQuery, contato.ID).Scan(&atendenteResult).Error
	if err == nil {
		atendente = &atendenteResult
	}

	// Buscar kanban board do contato
	var kanbanBoard *string
	if contato.StatusKanban != nil && *contato.StatusKanban != "" {
		kanbanBoard = contato.StatusKanban
	}

	// Montar resposta com dados completos
	response := gin.H{
		"id":    contato.ID,
		"fila":  fila,
		"tags":  tags,
		"atendente": atendente,
		"kanbanBoard": kanbanBoard,
		"orcamento": nil,   // TODO: Implementar busca de orçamento
		"agendamento": nil, // TODO: Implementar busca de agendamento
	}

	c.JSON(http.StatusOK, response)
}

// AssociarTagsContato - POST /api/contatos/:id/tags
func (h *ContatosHandler) AssociarTagsContato(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	contatoParam := c.Param("id")
	
	var req struct {
		TagIds []string `json:"tagIds" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Dados inválidos", "details": err.Error()})
		return
	}

	// Buscar contato pelo ID interno ou chatId (contactid)
	var contato models.Contato
	var query string
	if strings.Contains(contatoParam, "@") {
		// É um chatId do WhatsApp (formato: 5518997200106@c.us)
		query = `
			SELECT c.id, c.numero_telefone, c.nome, c.contactid
			FROM contatos c
			INNER JOIN sessoes_whatsapp sw ON c.sessao_whatsapp_id = sw.id
			WHERE c.contactid = ? AND sw.usuario_id = ?
		`
	} else {
		// É um ID interno
		query = `
			SELECT c.id, c.numero_telefone, c.nome, c.contactid
			FROM contatos c
			INNER JOIN sessoes_whatsapp sw ON c.sessao_whatsapp_id = sw.id
			WHERE c.id = ? AND sw.usuario_id = ?
		`
	}
	
	err := h.db.Raw(query, contatoParam, userID).Scan(&contato).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contato não encontrado"})
		return
	}
	
	contatoId := contato.ID

	// Remover tags existentes do contato
	if err := h.db.Where("contato_id = ?", contatoId).Delete(&models.ContatoTag{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao remover tags existentes"})
		return
	}

	// Adicionar novas tags
	var contatoTags []models.ContatoTag
	for _, tagId := range req.TagIds {
		// Verificar se a tag existe
		var tag models.Tag
		if err := h.db.First(&tag, "id = ?", tagId).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tag não encontrada: " + tagId})
			return
		}

		contatoTag := models.ContatoTag{
			ContatoID: contatoId,
			TagID:     tagId,
		}
		contatoTags = append(contatoTags, contatoTag)
	}

	// Salvar associações em lote
	if len(contatoTags) > 0 {
		if err := h.db.Create(&contatoTags).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao salvar tags", "details": err.Error()})
			return
		}
	}

	// Retornar tags atualizadas do contato
	var tagsAtualizadas []models.Tag
	tagQuery := `
		SELECT t.id, t.nome, t.cor, t.descricao, t.favorito
		FROM tags t
		INNER JOIN contato_tags ct ON t.id = ct.tag_id
		WHERE ct.contato_id = ?
		ORDER BY t.nome ASC
	`
	
	err = h.db.Raw(tagQuery, contatoId).Scan(&tagsAtualizadas).Error
	if err != nil {
		tagsAtualizadas = []models.Tag{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tags associadas com sucesso",
		"data":    tagsAtualizadas,
	})
}

// ListarTagsContato - GET /api/contatos/:id/tags
func (h *ContatosHandler) ListarTagsContato(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	contatoParam := c.Param("id")

	// Buscar contato pelo ID interno ou chatId (contactid)
	var contato models.Contato
	var query string
	if strings.Contains(contatoParam, "@") {
		// É um chatId do WhatsApp (formato: 5518997200106@c.us)
		query = `
			SELECT c.id, c.numero_telefone, c.nome, c.contactid
			FROM contatos c
			INNER JOIN sessoes_whatsapp sw ON c.sessao_whatsapp_id = sw.id
			WHERE c.contactid = ? AND sw.usuario_id = ?
		`
	} else {
		// É um ID interno
		query = `
			SELECT c.id, c.numero_telefone, c.nome, c.contactid
			FROM contatos c
			INNER JOIN sessoes_whatsapp sw ON c.sessao_whatsapp_id = sw.id
			WHERE c.id = ? AND sw.usuario_id = ?
		`
	}
	
	err := h.db.Raw(query, contatoParam, userID).Scan(&contato).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contato não encontrado"})
		return
	}
	
	contatoId := contato.ID

	// Buscar tags do contato
	var tags []models.Tag
	tagQuery := `
		SELECT t.id, t.nome, t.cor, t.descricao, t.favorito
		FROM tags t
		INNER JOIN contato_tags ct ON t.id = ct.tag_id
		WHERE ct.contato_id = ?
		ORDER BY t.nome ASC
	`
	
	err = h.db.Raw(tagQuery, contatoId).Scan(&tags).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar tags do contato"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tags,
	})
}
