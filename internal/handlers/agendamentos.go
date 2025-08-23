package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"tappyone/internal/models"
	"gorm.io/gorm"
)

// AgendamentosHandler gerencia os agendamentos
type AgendamentosHandler struct {
	db *gorm.DB
}

func NewAgendamentosHandler(db *gorm.DB) *AgendamentosHandler {
	return &AgendamentosHandler{db: db}
}

// ListAgendamentos lista todos os agendamentos do usuário
func (h *AgendamentosHandler) ListAgendamentos(c *gin.Context) {
	userID := c.GetString("user_id")
	contatoJID := c.Query("contato_id")
	status := c.Query("status")

	var agendamentos []models.Agendamento
	query := h.db.Where("usuario_id = ?", userID)
	
	if contatoJID != "" {
		// Converter JID para número de telefone e buscar contato
		numeroTelefone := strings.Replace(contatoJID, "@c.us", "", 1)
		
		// Buscar contato pelo número de telefone
		var contato models.Contato
		if err := h.db.Where("numero_telefone = ?", numeroTelefone).First(&contato).Error; err == nil {
			// Se contato encontrado, filtrar por UUID do contato
			query = query.Where("contato_id = ?", contato.ID)
		} else {
			// Se contato não encontrado, não retornar nenhum agendamento
			c.JSON(http.StatusOK, []models.Agendamento{})
			return
		}
	}
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	if err := query.Preload("Contato").Order("inicio_em ASC").Find(&agendamentos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar agendamentos"})
		return
	}

	// Debug: verificar se contatos estão sendo carregados
	log.Printf("Found %d agendamentos", len(agendamentos))
	for i, ag := range agendamentos {
		log.Printf("Agendamento %d: %s, ContatoID: %s, Contato loaded: %v", 
			i, ag.Titulo, ag.ContatoID, ag.Contato.ID != "")
	}

	c.JSON(http.StatusOK, agendamentos)
}

// GetAgendamento busca um agendamento específico
func (h *AgendamentosHandler) GetAgendamento(c *gin.Context) {
	userID := c.GetString("user_id")
	agendamentoID := c.Param("id")

	var agendamento models.Agendamento
	if err := h.db.Where("id = ? AND usuario_id = ?", agendamentoID, userID).First(&agendamento).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento não encontrado"})
		return
	}

	c.JSON(http.StatusOK, agendamento)
}

// CreateAgendamento cria um novo agendamento
func (h *AgendamentosHandler) CreateAgendamento(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		Titulo      string     `json:"titulo" binding:"required"`
		Descricao   *string    `json:"descricao"`
		InicioEm    time.Time  `json:"inicio_em" binding:"required"`
		FimEm       time.Time  `json:"fim_em" binding:"required"`
		LinkMeeting *string    `json:"link_meeting"`
		ContatoID   string     `json:"contato_id" binding:"required"`
		Contato     struct {
			ID       string  `json:"id"`
			Nome     *string `json:"nome"`
			Telefone *string `json:"telefone"`
			Email    *string `json:"email"`
			Empresa  *string `json:"empresa"`
			CPF      *string `json:"cpf"`
			CNPJ     *string `json:"cnpj"`
			CEP      *string `json:"cep"`
			Rua      *string `json:"rua"`
			Numero   *string `json:"numero"`
			Bairro   *string `json:"bairro"`
			Cidade   *string `json:"cidade"`
			Estado   *string `json:"estado"`
			Pais     *string `json:"pais"`
		} `json:"contato"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Debug: Log dos dados recebidos
	log.Printf("📝 Dados do contato recebidos:")
	log.Printf("  Nome: %v", req.Contato.Nome)
	log.Printf("  Email: %v", req.Contato.Email)
	log.Printf("  Empresa: %v", req.Contato.Empresa)
	log.Printf("  CPF: %v", req.Contato.CPF)
	log.Printf("  CNPJ: %v", req.Contato.CNPJ)
	log.Printf("  CEP: %v", req.Contato.CEP)

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
		
		// Se contato não existe, criar um novo com todos os dados
		contato = models.Contato{
			NumeroTelefone:   numeroTelefone,
			Nome:             req.Contato.Nome,
			Email:            req.Contato.Email,
			Empresa:          req.Contato.Empresa,
			CPF:              req.Contato.CPF,
			CNPJ:             req.Contato.CNPJ,
			CEP:              req.Contato.CEP,
			Rua:              req.Contato.Rua,
			Numero:           req.Contato.Numero,
			Bairro:           req.Contato.Bairro,
			Cidade:           req.Contato.Cidade,
			Estado:           req.Contato.Estado,
			Pais:             req.Contato.Pais,
			SessaoWhatsappID: sessaoWhatsapp.ID,
		}
		if err := h.db.Create(&contato).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar contato"})
			return
		}
	} else {
		// Se contato existe, atualizar com novos dados
		updates := map[string]interface{}{}
		
		if req.Contato.Nome != nil && *req.Contato.Nome != "" {
			updates["nome"] = req.Contato.Nome
		}
		if req.Contato.Email != nil && *req.Contato.Email != "" {
			updates["email"] = req.Contato.Email
		}
		if req.Contato.Empresa != nil && *req.Contato.Empresa != "" {
			updates["empresa"] = req.Contato.Empresa
		}
		if req.Contato.CPF != nil && *req.Contato.CPF != "" {
			updates["cpf"] = req.Contato.CPF
		}
		if req.Contato.CNPJ != nil && *req.Contato.CNPJ != "" {
			updates["cnpj"] = req.Contato.CNPJ
		}
		if req.Contato.CEP != nil && *req.Contato.CEP != "" {
			updates["cep"] = req.Contato.CEP
		}
		if req.Contato.Rua != nil && *req.Contato.Rua != "" {
			updates["rua"] = req.Contato.Rua
		}
		if req.Contato.Numero != nil && *req.Contato.Numero != "" {
			updates["numero"] = req.Contato.Numero
		}
		if req.Contato.Bairro != nil && *req.Contato.Bairro != "" {
			updates["bairro"] = req.Contato.Bairro
		}
		if req.Contato.Cidade != nil && *req.Contato.Cidade != "" {
			updates["cidade"] = req.Contato.Cidade
		}
		if req.Contato.Estado != nil && *req.Contato.Estado != "" {
			updates["estado"] = req.Contato.Estado
		}
		if req.Contato.Pais != nil && *req.Contato.Pais != "" {
			updates["pais"] = req.Contato.Pais
		}
		
		if len(updates) > 0 {
			if err := h.db.Model(&contato).Updates(updates).Error; err != nil {
				log.Printf("Erro ao atualizar contato: %v", err)
			}
		}
	}

	agendamento := models.Agendamento{
		Titulo:      req.Titulo,
		Descricao:   req.Descricao,
		InicioEm:    req.InicioEm,
		FimEm:       req.FimEm,
		LinkMeeting: req.LinkMeeting,
		Status:      "AGENDADO",
		UsuarioID:   userID,
		ContatoID:   contato.ID, // Usar ID do contato encontrado/criado
	}

	if err := h.db.Create(&agendamento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar agendamento"})
		return
	}

	// Recarregar agendamento com contato completo
	if err := h.db.Preload("Contato").Where("id = ?", agendamento.ID).First(&agendamento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao carregar agendamento criado"})
		return
	}

	c.JSON(http.StatusCreated, agendamento)
}

// UpdateAgendamento atualiza um agendamento existente
func (h *AgendamentosHandler) UpdateAgendamento(c *gin.Context) {
	userID := c.GetString("user_id")
	agendamentoID := c.Param("id")

	var req struct {
		Titulo      *string    `json:"titulo"`
		Descricao   *string    `json:"descricao"`
		DataHora    *time.Time `json:"data_hora"`
		InicioEm    *time.Time `json:"inicioEm"`
		FimEm       *time.Time `json:"fimEm"`
		Duracao     *int       `json:"duracao"`
		Local       *string    `json:"local"`
		LinkMeeting *string    `json:"link_meeting"`
		Status      *string    `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var agendamento models.Agendamento
	if err := h.db.Where("id = ? AND usuario_id = ?", agendamentoID, userID).First(&agendamento).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento não encontrado"})
		return
	}

	// Atualizar campos fornecidos
	updates := make(map[string]interface{})
	if req.Titulo != nil {
		updates["titulo"] = *req.Titulo
	}
	if req.Descricao != nil {
		updates["descricao"] = *req.Descricao
	}
	if req.DataHora != nil {
		updates["data_hora"] = *req.DataHora
	}
	if req.InicioEm != nil {
		updates["inicio_em"] = *req.InicioEm
	}
	if req.FimEm != nil {
		updates["fim_em"] = *req.FimEm
	}
	if req.Duracao != nil {
		updates["duracao"] = *req.Duracao
	}
	if req.Local != nil {
		updates["local"] = *req.Local
	}
	if req.LinkMeeting != nil {
		updates["link_meeting"] = *req.LinkMeeting
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	updates["atualizado_em"] = time.Now()

	if err := h.db.Model(&agendamento).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar agendamento"})
		return
	}

	// Recarregar dados atualizados
	if err := h.db.Where("id = ?", agendamentoID).First(&agendamento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao recarregar agendamento"})
		return
	}

	c.JSON(http.StatusOK, agendamento)
}

// DeleteAgendamento exclui um agendamento
func (h *AgendamentosHandler) DeleteAgendamento(c *gin.Context) {
	userID := c.GetString("user_id")
	agendamentoID := c.Param("id")

	var agendamento models.Agendamento
	if err := h.db.Where("id = ? AND usuario_id = ?", agendamentoID, userID).First(&agendamento).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento não encontrado"})
		return
	}

	if err := h.db.Delete(&agendamento).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao excluir agendamento"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agendamento excluído com sucesso"})
}

// GetAgendamentosByContato busca todos os agendamentos de um contato específico
func (h *AgendamentosHandler) GetAgendamentosByContato(c *gin.Context) {
	userID := c.GetString("user_id")
	contatoID := c.Param("contato_id")

	var agendamentos []models.Agendamento
	if err := h.db.Where("usuario_id = ? AND contato_id = ?", userID, contatoID).Order("data_hora ASC").Find(&agendamentos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar agendamentos do contato"})
		return
	}

	c.JSON(http.StatusOK, agendamentos)
}

// UpdateAgendamentoStatus atualiza apenas o status do agendamento
func (h *AgendamentosHandler) UpdateAgendamentoStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	agendamentoID := c.Param("id")

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var agendamento models.Agendamento
	if err := h.db.Where("id = ? AND usuario_id = ?", agendamentoID, userID).First(&agendamento).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agendamento não encontrado"})
		return
	}

	if err := h.db.Model(&agendamento).Updates(map[string]interface{}{
		"status": req.Status,
		"atualizado_em": time.Now(),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status atualizado com sucesso"})
}

// GetAgendamentosProximos busca agendamentos próximos (próximas 24 horas)
func (h *AgendamentosHandler) GetAgendamentosProximos(c *gin.Context) {
	userID := c.GetString("user_id")

	agora := time.Now()
	proximasDias := agora.Add(24 * time.Hour)

	var agendamentos []models.Agendamento
	if err := h.db.Where("usuario_id = ? AND inicio_em BETWEEN ? AND ? AND status = ?", 
		userID, agora, proximasDias, "AGENDADO").
		Order("inicio_em ASC").Find(&agendamentos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar agendamentos próximos"})
		return
	}

	c.JSON(http.StatusOK, agendamentos)
}
