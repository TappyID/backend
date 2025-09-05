package router

import (
	"fmt"
	"log"
	"strconv"
	"tappyone/internal/handlers"
	"tappyone/internal/middleware"
	"tappyone/internal/services"
	"tappyone/internal/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Setup configura todas as rotas da aplicação
func Setup(container *services.Container) *gin.Engine {
	r := gin.Default()

	// Inicializar WebSocket Hub
	handlers.InitWebSocketHub()

	// CORS - habilitado sempre para desenvolvimento
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"https://crm.tappy.id",
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization", "x-api-key"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// Middleware de logging
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Inicializar handlers
	log.Printf("[ROUTER] Inicializando handlers...")
	authHandler := handlers.NewAuthHandler(container.AuthService)
	userHandler := handlers.NewUserHandler(container.UserService)
	contatoHandler := handlers.NewContatosHandler(container.DB)
	kanbanHandler := handlers.NewKanbanHandler(container.KanbanService)
	agendamentoHandler := handlers.NewAgendamentosHandler(container.DB)
	log.Printf("[ROUTER] AgendamentosHandler criado: %v", agendamentoHandler != nil)
	orcamentoHandler := handlers.NewOrcamentosHandler(container.DB)
	whatsAppHandler := handlers.NewWhatsAppHandler(container.WhatsAppService)
	fluxosHandler := handlers.NewFluxosHandler(container.DB, container.FluxoExecutionService)
	respostaRapidaHandler := handlers.NewRespostaRapidaHandler(container.RespostaRapidaService)
	connectionHandler := handlers.NewConnectionHandler(container.ConnectionService)
	anotacoesHandler := handlers.NewAnotacoesHandler(container.DB)
	log.Printf("[ROUTER] AnotacoesHandler criado: %v", anotacoesHandler != nil)
	assinaturasHandler := handlers.NewAssinaturasHandler(container.DB)
	log.Printf("[ROUTER] AssinaturasHandler criado: %v", assinaturasHandler != nil)
	whatsappMediaHandler := handlers.NewWhatsAppMediaHandler(container.WhatsAppService, container.AuthService)
	filasHandler := handlers.NewFilasHandler(container.DB)
	tagsHandler := handlers.NewTagsHandler(container.DB, container.AuthService)
	alertasHandler := handlers.NewAlertasHandler(container.DB, container.AuthService)
	atendimentoStatsHandler := handlers.NewAtendimentoStatsHandler(container.WhatsAppService, container.DB)
	sessoesWhatsAppHandler := handlers.NewSessoesWhatsAppHandler(container.DB)
	log.Printf("[ROUTER] Todos os handlers criados com sucesso")

	// Servir arquivos estáticos (uploads)
	r.Static("/uploads", "./uploads")

	// Rotas públicas
	public := r.Group("/api")
	{
		public.POST("/auth/login", authHandler.Login)
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "message": "TappyOne CRM API"})
		})
		public.GET("/version", func(c *gin.Context) {
			c.JSON(200, gin.H{"version": "1.0.0", "service": "tappyone-backend"})
		})
	}

	// Rota de teste simples (sem grupo)
	r.GET("/test", func(c *gin.Context) {
		log.Printf("[TEST] Route /test accessed")
		c.JSON(200, gin.H{"message": "Backend is working!", "timestamp": "2025-01-20"})
	})

	// WebSocket route (with JWT auth via query param)
	r.GET("/ws", handlers.NewWebSocketHandler(container.AuthService, container.Config.JWTSecret))

	// Webhook específico para respostas rápidas
	r.POST("/webhooks/resposta-rapida", respostaRapidaHandler.ProcessarMensagemWebhook)

	// Rotas protegidas
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware(container.AuthService))
	{
		// Auth
		protected.GET("/auth/me", authHandler.Me)

		// Usuários
		users := protected.Group("/users")
		{
			users.GET("/me", userHandler.GetMe)
			users.PUT("/me", userHandler.UpdateMe)
			users.GET("/", userHandler.List)
			users.POST("/", userHandler.Create)
			users.GET("/:id", userHandler.GetByID)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}

		// WhatsApp
		whatsapp := protected.Group("/whatsapp")
		{
			whatsapp.POST("/sessions", whatsAppHandler.CreateSession)
			whatsapp.GET("/sessions", whatsAppHandler.ListSessions)
			whatsapp.GET("/sessions/:id", whatsAppHandler.GetSession)
			whatsapp.POST("/sessions/:id/start", whatsAppHandler.StartSession)
			whatsapp.POST("/sessions/:id/stop", whatsAppHandler.StopSession)
			whatsapp.GET("/sessions/:id/qr", whatsAppHandler.GetQRCode)
		}

		// Kanban
		kanban := protected.Group("/kanban")
		{
			kanban.GET("/quadros", kanbanHandler.ListQuadros)
			kanban.POST("/quadros", kanbanHandler.CreateQuadro)
			kanban.GET("/quadros/:id", kanbanHandler.GetQuadro)
			kanban.PUT("/quadros/:id", kanbanHandler.UpdateQuadro)
			kanban.DELETE("/quadros/:id", kanbanHandler.DeleteQuadro)

			// Colunas
			kanban.POST("/column-create", kanbanHandler.CreateColumn)
			kanban.POST("/column-edit", kanbanHandler.EditColumn)
			kanban.POST("/column-delete", kanbanHandler.DeleteColumn)
			kanban.PUT("/coluna/:colunaId/color", kanbanHandler.UpdateColumnColor)
			kanban.PUT("/coluna/reorder", kanbanHandler.ReorderColumns)
			kanban.POST("/card-movement", kanbanHandler.MoveCard)
			kanban.GET("/:id/metadata", kanbanHandler.GetMetadata)
		}

		// Connections
		connections := protected.Group("/connections")
		{
			connections.GET("/", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Connections"})
			})
			connections.GET("/whatsapp", connectionHandler.GetWhatsAppConnection)
			connections.POST("/whatsapp", connectionHandler.CreateOrUpdateConnection)
			connections.POST("/whatsapp/sync/:sessionName", connectionHandler.SyncWhatsAppConnection)
			connections.DELETE("/whatsapp/:sessionName", connectionHandler.DisconnectWhatsApp)
		}

		// Respostas Rápidas
		respostasRapidas := protected.Group("/respostas-rapidas")
		{
			// Categorias
			respostasRapidas.GET("/categorias", respostaRapidaHandler.GetCategorias)
			respostasRapidas.POST("/categorias", respostaRapidaHandler.CreateCategoria)
			respostasRapidas.PUT("/categorias/:id", respostaRapidaHandler.UpdateCategoria)
			respostasRapidas.DELETE("/categorias/:id", respostaRapidaHandler.DeleteCategoria)

			// Execuções e Estatísticas (antes das rotas com :id)
			respostasRapidas.GET("/execucoes", respostaRapidaHandler.GetExecucoes)
			respostasRapidas.GET("/estatisticas", respostaRapidaHandler.GetEstatisticas)

			// Agendamentos (antes das rotas com :id)
			respostasRapidas.GET("/agendamentos", respostaRapidaHandler.GetAgendamentos)
			respostasRapidas.PUT("/agendamentos/:id/pausar", respostaRapidaHandler.PausarAgendamento)

			// Comandos Slash (antes das rotas com :id)
			respostasRapidas.POST("/comando-slash", respostaRapidaHandler.ProcessarComandoSlash)

			// Respostas Rápidas
			respostasRapidas.GET("/", respostaRapidaHandler.GetRespostasRapidas)
			respostasRapidas.POST("/", respostaRapidaHandler.CreateRespostaRapida)
			respostasRapidas.GET("/:id", respostaRapidaHandler.GetRespostaRapida)
			respostasRapidas.PUT("/:id", respostaRapidaHandler.UpdateRespostaRapida)
			respostasRapidas.DELETE("/:id", respostaRapidaHandler.DeleteRespostaRapida)
			respostasRapidas.PUT("/:id/pausar", respostaRapidaHandler.TogglePausarRespostaRapida)
			respostasRapidas.POST("/:id/executar", respostaRapidaHandler.ExecutarRespostaRapida)
			respostasRapidas.GET("/:id/acoes", respostaRapidaHandler.GetAcoes)
			respostasRapidas.POST("/:id/acoes", respostaRapidaHandler.CreateAcao)
			respostasRapidas.PUT("/:id/acoes/reorder", respostaRapidaHandler.ReorderAcoes)
		}

		// Anotações
		log.Printf("[ROUTER] Registrando rotas de anotações...")
		anotacoes := protected.Group("/anotacoes")
		{
			anotacoes.GET("", anotacoesHandler.ListAnotacoes)
			anotacoes.GET("/:id", anotacoesHandler.GetAnotacao)
			anotacoes.POST("", anotacoesHandler.CreateAnotacao)
			anotacoes.PUT("/:id", anotacoesHandler.UpdateAnotacao)
			anotacoes.DELETE("/:id", anotacoesHandler.DeleteAnotacao)
			anotacoes.GET("/contato/:contato_id", anotacoesHandler.GetAnotacoesByContato)
		}
		log.Printf("[ROUTER] Rotas de anotações registradas")

		// Agendamentos
		log.Printf("[ROUTER] Registrando rotas de agendamentos...")
		agendamentos := protected.Group("/agendamentos")
		{
			agendamentos.GET("", agendamentoHandler.ListAgendamentos)
			agendamentos.GET("/:id", agendamentoHandler.GetAgendamento)
			agendamentos.POST("", agendamentoHandler.CreateAgendamento)
			agendamentos.PUT("/:id", agendamentoHandler.UpdateAgendamento)
			agendamentos.DELETE("/:id", agendamentoHandler.DeleteAgendamento)
			agendamentos.GET("/proximos", agendamentoHandler.GetAgendamentosProximos)
		}
		log.Printf("[ROUTER] Rotas de agendamentos registradas")

		// Orçamentos
		orcamentos := protected.Group("/orcamentos")
		{
			orcamentos.GET("", orcamentoHandler.ListOrcamentos)
			orcamentos.GET("/:id", orcamentoHandler.GetOrcamento)
			orcamentos.POST("", orcamentoHandler.CreateOrcamento)
			orcamentos.PUT("/:id", orcamentoHandler.UpdateOrcamento)
			orcamentos.DELETE("/:id", orcamentoHandler.DeleteOrcamento)
		}

		// Assinaturas
		log.Printf("[ROUTER] Registrando rotas de assinaturas...")
		assinaturas := protected.Group("/assinaturas")
		{
			assinaturas.GET("/", assinaturasHandler.ListAssinaturas)
			assinaturas.POST("/", assinaturasHandler.CreateAssinatura)
			assinaturas.GET("/:id", assinaturasHandler.GetAssinatura)
			assinaturas.PUT("/:id", assinaturasHandler.UpdateAssinatura)
			assinaturas.DELETE("/:id", assinaturasHandler.DeleteAssinatura)
		}
		log.Printf("[ROUTER] Rotas de assinaturas registradas")

		// Filas
		filas := protected.Group("/filas")
		{
			filas.GET("/", filasHandler.ListarFilas)
			filas.POST("/", filasHandler.CriarFila)
			filas.GET("/:id", filasHandler.ObterFila)
			filas.PUT("/:id", filasHandler.AtualizarFila)
			filas.DELETE("/:id", filasHandler.DeletarFila)
			filas.PATCH("/:id/toggle", filasHandler.ToggleFilaStatus)
			filas.POST("/reordenar", filasHandler.ReordenarFilas)
			filas.POST("/:id/duplicar", filasHandler.DuplicarFila)
		}

		// Tags
		tags := protected.Group("/tags")
		{
			tags.GET("/", tagsHandler.ListarTags)
			tags.POST("/", tagsHandler.CriarTag)
			tags.GET("/favoritas", tagsHandler.ListarTagsFavoritas)
			tags.GET("/populares", tagsHandler.ListarTagsPopulares)
			tags.GET("/:id", tagsHandler.ObterTag)
			tags.PUT("/:id", tagsHandler.AtualizarTag)
			tags.DELETE("/:id", tagsHandler.DeletarTag)
		}

		// Alertas
		alertas := protected.Group("/alertas")
		{
			alertas.GET("", alertasHandler.ListarAlertas)
			alertas.GET("/:id", alertasHandler.ObterAlerta)
			alertas.POST("", alertasHandler.CriarAlerta)
			alertas.PUT("/:id", alertasHandler.AtualizarAlerta)
			alertas.DELETE("/:id", alertasHandler.DeletarAlerta)
		}

		// Estatísticas de Atendimento
		atendimentos := protected.Group("/atendimentos")
		{
			atendimentos.GET("/stats", atendimentoStatsHandler.GetStats)
		}

		// Contatos
		contatos := protected.Group("/contatos")
		{
			contatos.GET("", contatoHandler.ListContatos)
			contatos.GET("/stats", contatoHandler.GetContatosStats)
			contatos.GET("/export", contatoHandler.ExportContatos)
			contatos.POST("/import", contatoHandler.ImportContatos)
			contatos.GET("/:id", contatoHandler.GetContato)
			contatos.POST("", contatoHandler.CreateContato)
			contatos.PUT("/:id", contatoHandler.UpdateContato)
			contatos.DELETE("/:id", contatoHandler.DeleteContato)
			contatos.POST("/sync", contatoHandler.SyncContatos)
			// Buscar dados completos por chatId do WAHA
			contatos.GET("/chat/:chatId/dados-completos", contatoHandler.GetContatoDadosCompletos)
			// Tags do contato
			contatos.GET("/:id/tags", contatoHandler.ListarTagsContato)
			contatos.POST("/:id/tags", contatoHandler.AssociarTagsContato)
		}

		// Sessões WhatsApp
		sessoesWhatsApp := protected.Group("/sessoes-whatsapp")
		{
			sessoesWhatsApp.GET("", sessoesWhatsAppHandler.ListSessoesWhatsApp)
			sessoesWhatsApp.GET("/:id", sessoesWhatsAppHandler.GetSessaoWhatsApp)
			sessoesWhatsApp.POST("", sessoesWhatsAppHandler.CreateSessaoWhatsApp)
			sessoesWhatsApp.PUT("/:id", sessoesWhatsAppHandler.UpdateSessaoWhatsApp)
			sessoesWhatsApp.DELETE("/:id", sessoesWhatsAppHandler.DeleteSessaoWhatsApp)
		}

		// Agentes IA
		agentesHandler := handlers.NewAgentesHandler(container.DB)
		agentes := protected.Group("/agentes")
		{
			agentes.GET("", agentesHandler.GetAgentes)
			agentes.POST("", agentesHandler.CreateAgente)
			agentes.GET("/ativos", agentesHandler.GetAgentesAtivos)
			agentes.GET("/:id", agentesHandler.GetAgente)
			agentes.PUT("/:id", agentesHandler.UpdateAgente)
			agentes.DELETE("/:id", agentesHandler.DeleteAgente)
			agentes.POST("/:id/toggle", agentesHandler.ToggleAgente)
		}

		// Chat Agentes
		chatAgentes := protected.Group("/chat-agentes")
		{
			chatAgentes.GET("/:chatId", agentesHandler.GetChatAgente)
			chatAgentes.POST("/:chatId/activate", agentesHandler.ActivateAgentForChat)
			chatAgentes.POST("/:chatId/deactivate", agentesHandler.DeactivateAgentForChat)
		}

		// Fluxos (Automation Workflows)
		fluxos := protected.Group("/fluxos")
		{
			fluxos.GET("", fluxosHandler.ListFluxos)
			fluxos.POST("", fluxosHandler.CreateFluxo)
			fluxos.GET("/stats", fluxosHandler.GetFluxosStats)
			fluxos.GET("/:id", fluxosHandler.GetFluxo)
			fluxos.PUT("/:id", fluxosHandler.UpdateFluxo)
			fluxos.DELETE("/:id", fluxosHandler.DeleteFluxo)
			fluxos.PUT("/:id/toggle", fluxosHandler.ToggleFluxo)
			fluxos.POST("/:id/execute", fluxosHandler.ExecuteFluxo)

			// Nodes
			fluxos.POST("/:id/nodes", fluxosHandler.CreateFluxoNo)
			fluxos.PUT("/:id/nodes/:nodeId", fluxosHandler.UpdateFluxoNo)
			fluxos.DELETE("/:id/nodes/:nodeId", fluxosHandler.DeleteFluxoNo)

			// Connections
			fluxos.POST("/:id/connections", fluxosHandler.CreateFluxoConexao)
			fluxos.DELETE("/:id/connections/:connectionId", fluxosHandler.DeleteFluxoConexao)
		}

		// WhatsApp API (com middleware JWT)
		whatsappAPI := protected.Group("/whatsapp")
		{
			whatsappAPI.GET("/chats", func(c *gin.Context) {
				log.Printf("[WHATSAPP] GET /chats - Starting request")

				userID, exists := c.Get("userID")
				if !exists {
					c.JSON(401, gin.H{"error": "User not authenticated"})
					return
				}

				log.Printf("[WHATSAPP] GET /chats - UserID: %s, SessionName: user_%s", userID, userID)
				sessionName := fmt.Sprintf("user_%s", userID)

				chats, err := container.WhatsAppService.GetChats(sessionName)
				if err != nil {
					log.Printf("[WHATSAPP] GET /chats - Error: %v", err)
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}

				c.JSON(200, chats)
			})

			whatsappAPI.GET("/contacts", func(c *gin.Context) {
				userID, exists := c.Get("userID")
				if !exists {
					c.JSON(401, gin.H{"error": "User not authenticated"})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)

				var contacts interface{}
				contacts, err := container.WhatsAppService.GetContacts(sessionName)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, contacts)
			})

			whatsappAPI.GET("/groups", func(c *gin.Context) {
				log.Printf("[WHATSAPP] GET /groups - Starting request")

				userID, exists := c.Get("userID")
				if !exists {
					c.JSON(401, gin.H{"error": "User not authenticated"})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)
				log.Printf("[WHATSAPP] GET /groups - UserID: %s, SessionName: %s", userID, sessionName)

				var groups interface{}
				groups, err := container.WhatsAppService.GetGroups(sessionName)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, groups)
			})

			whatsappAPI.GET("/chats/:chatId/messages", func(c *gin.Context) {
				// Autenticação via header Authorization
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					c.JSON(401, gin.H{"error": "Invalid token", "details": err.Error()})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)
				chatID := c.Param("chatId")

				var messages interface{}
				// Parse query parameters for pagination
				limitStr := c.DefaultQuery("limit", "50")
				offsetStr := c.DefaultQuery("offset", "0")

				limit, err := strconv.Atoi(limitStr)
				if err != nil {
					limit = 50
				}

				offset, err := strconv.Atoi(offsetStr)
				if err != nil {
					offset = 0
				}

				messages, err = container.WhatsAppService.GetChatMessages(sessionName, chatID, limit, offset)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, messages)
			})

			whatsappAPI.POST("/chats/:chatId/messages", func(c *gin.Context) {
				// Autenticação via header Authorization
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					c.JSON(401, gin.H{"error": "Invalid token", "details": err.Error()})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)
				chatID := c.Param("chatId")

				var req struct {
					Text     string   `json:"text" binding:"required"`
					ReplyTo  string   `json:"replyTo,omitempty"`
					Mentions []string `json:"mentions,omitempty"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": "Invalid request"})
					return
				}

				var result interface{}
				if req.ReplyTo != "" {
					result, err = container.WhatsAppService.SendReplyMessage(sessionName, chatID, req.Text, req.ReplyTo)
				} else if len(req.Mentions) > 0 {
					result, err = container.WhatsAppService.SendMessageWithMentions(sessionName, chatID, req.Text, req.Mentions)
				} else {
					result, err = container.WhatsAppService.SendMessage(sessionName, chatID, req.Text)
				}

				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

			// Endpoint para marcar mensagens como lidas
			whatsappAPI.POST("/chats/:chatId/read", func(c *gin.Context) {
				userID, exists := c.Get("userID")
				if !exists {
					c.JSON(401, gin.H{"error": "User not authenticated"})
					return
				}

				sessionName := fmt.Sprintf("user_%s", userID)
				chatID := c.Param("chatId")

				result, err := container.WhatsAppService.MarkAsRead(sessionName, chatID)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

			// Anti-blocking endpoints conforme WAHA best practices
			// 1. Marcar como visualizada (sendSeen)
			whatsappAPI.POST("/chats/:chatId/seen", func(c *gin.Context) {
				userID, exists := c.Get("userID")
				if !exists {
					c.JSON(401, gin.H{"error": "User not authenticated"})
					return
				}

				sessionName := fmt.Sprintf("user_%s", userID)
				chatID := c.Param("chatId")

				result, err := container.WhatsAppService.SendSeenAntiBlock(sessionName, chatID)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

			// 2. Começar a digitar (startTyping)
			whatsappAPI.POST("/chats/:chatId/typing/start", func(c *gin.Context) {
				userID, exists := c.Get("userID")
				if !exists {
					c.JSON(401, gin.H{"error": "User not authenticated"})
					return
				}

				sessionName := fmt.Sprintf("user_%s", userID)
				chatID := c.Param("chatId")

				result, err := container.WhatsAppService.StartTyping(sessionName, chatID)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

			// 3. Parar de digitar (stopTyping)
			whatsappAPI.POST("/chats/:chatId/typing/stop", func(c *gin.Context) {
				userID, exists := c.Get("userID")
				if !exists {
					c.JSON(401, gin.H{"error": "User not authenticated"})
					return
				}

				sessionName := fmt.Sprintf("user_%s", userID)
				chatID := c.Param("chatId")

				result, err := container.WhatsAppService.StopTyping(sessionName, chatID)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

			// Rotas de mídia - usando handlers que esperam JSON com URLs
			whatsappAPI.POST("/chats/:chatId/voice", whatsappMediaHandler.SendVoice)
			whatsappAPI.POST("/chats/:chatId/image", whatsAppHandler.SendImageMessage)
			whatsappAPI.POST("/chats/:chatId/video", whatsappMediaHandler.SendVideoMessage)
			whatsappAPI.POST("/chats/:chatId/file", whatsappMediaHandler.SendFile)
			whatsappAPI.POST("/chats/:chatId/contact", whatsappMediaHandler.SendContact)
			whatsappAPI.GET("/media/:mediaId", whatsAppHandler.DownloadMedia)

			// Rota de upload e envio de mídia combinado
			whatsappAPI.POST("/upload-and-send-media", whatsappMediaHandler.UploadAndSendMedia)

			// Reações
			whatsappAPI.PUT("/messages/:messageId/reaction", func(c *gin.Context) {
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					c.JSON(401, gin.H{"error": "Invalid token"})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)
				messageID := c.Param("messageId")

				var req struct {
					Reaction string `json:"reaction"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": "Invalid request"})
					return
				}

				if req.Reaction == "" {
					err = container.WhatsAppService.RemoveReaction(sessionName, messageID)
				} else {
					err = container.WhatsAppService.AddReaction(sessionName, messageID, req.Reaction)
				}

				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, gin.H{"success": true})
			})

			// Encaminhar mensagens
			whatsappAPI.POST("/messages/:messageId/forward", func(c *gin.Context) {
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					c.JSON(401, gin.H{"error": "Invalid token"})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)
				messageID := c.Param("messageId")

				var req struct {
					ToChatID string `json:"toChatId" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": "Invalid request"})
					return
				}

				result, err := container.WhatsAppService.ForwardMessage(sessionName, req.ToChatID, messageID)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

			// Editar mensagens
			whatsappAPI.PUT("/chats/:chatId/messages/:messageId", func(c *gin.Context) {
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					c.JSON(401, gin.H{"error": "Invalid token"})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)
				chatID := c.Param("chatId")
				messageID := c.Param("messageId")

				var req struct {
					Text string `json:"text" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": "Invalid request"})
					return
				}

				result, err := container.WhatsAppService.EditMessage(sessionName, chatID, messageID, req.Text)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

			// Deletar mensagens
			whatsappAPI.DELETE("/chats/:chatId/messages/:messageId", func(c *gin.Context) {
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					c.JSON(401, gin.H{"error": "Invalid token"})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)
				chatID := c.Param("chatId")
				messageID := c.Param("messageId")

				result, err := container.WhatsAppService.DeleteMessage(sessionName, chatID, messageID)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

			// Arquivar chat
			whatsappAPI.POST("/chats/:chatId/archive", func(c *gin.Context) {
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					c.JSON(401, gin.H{"error": "Invalid token"})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)
				chatID := c.Param("chatId")

				result, err := container.WhatsAppService.ArchiveChat(sessionName, chatID)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

			// Desarquivar chat
			whatsappAPI.POST("/chats/:chatId/unarchive", func(c *gin.Context) {
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					c.JSON(401, gin.H{"error": "Invalid token"})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)
				chatID := c.Param("chatId")

				result, err := container.WhatsAppService.UnarchiveChat(sessionName, chatID)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

			// Deletar chat completo
			whatsappAPI.DELETE("/chats/:chatId", func(c *gin.Context) {
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					c.JSON(401, gin.H{"error": "Invalid token"})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)
				chatID := c.Param("chatId")

				result, err := container.WhatsAppService.DeleteChat(sessionName, chatID)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

			// Favoritar mensagens
			whatsappAPI.PUT("/star", func(c *gin.Context) {
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					c.JSON(401, gin.H{"error": "Invalid token"})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)

				var req struct {
					MessageID string `json:"messageId" binding:"required"`
					Star      bool   `json:"star"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": "Invalid request"})
					return
				}

				result, err := container.WhatsAppService.StarMessage(sessionName, req.MessageID, req.Star)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

			// Enviar contato
			whatsappAPI.POST("/sendContactVcard", func(c *gin.Context) {
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					c.JSON(401, gin.H{"error": "Invalid token"})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)

				var req struct {
					ChatID    string `json:"chatId" binding:"required"`
					ContactID string `json:"contactId" binding:"required"`
					Name      string `json:"name" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": "Invalid request"})
					return
				}

				result, err := container.WhatsAppService.SendContactVcard(sessionName, req.ChatID, req.ContactID, req.Name)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

			// Enviar localização
			whatsappAPI.POST("/sendLocation", func(c *gin.Context) {
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					c.JSON(401, gin.H{"error": "Invalid token"})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)

				var req struct {
					ChatID    string  `json:"chatId" binding:"required"`
					Latitude  float64 `json:"latitude" binding:"required"`
					Longitude float64 `json:"longitude" binding:"required"`
					Title     string  `json:"title"`
					Address   string  `json:"address"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": "Invalid request"})
					return
				}

				result, err := container.WhatsAppService.SendLocation(sessionName, req.ChatID, req.Latitude, req.Longitude, req.Title, req.Address)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

			// Enviar enquete
			whatsappAPI.POST("/sendPoll", func(c *gin.Context) {
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					c.JSON(401, gin.H{"error": "Invalid token"})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)

				var req struct {
					ChatID          string   `json:"chatId" binding:"required"`
					Name            string   `json:"name" binding:"required"`
					Options         []string `json:"options" binding:"required"`
					MultipleAnswers bool     `json:"multipleAnswers"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": "Invalid request"})
					return
				}

				result, err := container.WhatsAppService.SendPoll(sessionName, req.ChatID, req.Name, req.Options, req.MultipleAnswers)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

			// Buscar mensagens
			whatsappAPI.GET("/chats/:chatId/messages/search", func(c *gin.Context) {
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					c.JSON(401, gin.H{"error": "Invalid token"})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)
				chatID := c.Param("chatId")
				query := c.Query("q")

				if query == "" {
					c.JSON(400, gin.H{"error": "Query parameter 'q' is required"})
					return
				}

				// Parse pagination parameters
				limitStr := c.DefaultQuery("limit", "50")
				offsetStr := c.DefaultQuery("offset", "0")

				limit, err := strconv.Atoi(limitStr)
				if err != nil {
					limit = 50
				}

				offset, err := strconv.Atoi(offsetStr)
				if err != nil {
					offset = 0
				}

				result, err := container.WhatsAppService.SearchMessages(sessionName, chatID, query, limit, offset)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})

		}

		// Reply endpoint
		whatsappAPI.POST("/reply", func(c *gin.Context) {
			userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
			if err != nil {
				c.JSON(401, gin.H{"error": "Invalid token"})
				return
			}
			sessionName := fmt.Sprintf("user_%s", userID)

			var req struct {
				ChatID  string `json:"chatId" binding:"required"`
				Text    string `json:"text" binding:"required"`
				ReplyTo string `json:"replyTo" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": "Invalid request"})
				return
			}

			result, err := container.WhatsAppService.SendReplyMessage(sessionName, req.ChatID, req.Text, req.ReplyTo)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, result)
		})

		// SendSeen endpoint
		whatsappAPI.POST("/sendSeen", func(c *gin.Context) {
			userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
			if err != nil {
				c.JSON(401, gin.H{"error": "Invalid token"})
				return
			}
			sessionName := fmt.Sprintf("user_%s", userID)

			var req struct {
				ChatID     string   `json:"chatId" binding:"required"`
				MessageIDs []string `json:"messageIds" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": "Invalid request"})
				return
			}

			err = container.WhatsAppService.SendSeen(sessionName, req.ChatID, req.MessageIDs)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"success": true})
		})
	}

	// Send endpoints
	send := protected.Group("/send")
	{
		// Link preview endpoint
		send.POST("/link-custom-preview", func(c *gin.Context) {
			userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
			if err != nil {
				c.JSON(401, gin.H{"error": "Invalid token"})
				return
			}

			var req struct {
				URL string `json:"url" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": "Invalid request"})
				return
			}

			// Implementação básica de link preview
			// Por enquanto, retorna dados básicos extraídos da URL
			result := gin.H{
				"title":       "Link Preview",
				"description": "Preview do link: " + req.URL,
				"image":       "",
				"favicon":     "",
				"url":         req.URL,
				"siteName":    "",
			}

			log.Printf("[LINK_PREVIEW] User %s requested preview for: %s", userID, req.URL)
			c.JSON(200, result)
		})
	}

	// WAHA API Proxy routes
	wahaProxy := r.Group("/api")
	{
		// QR Code routes
		wahaProxy.GET("/:session/auth/qr", whatsAppHandler.ProxyToWAHA)
		wahaProxy.GET("/:session/qr", whatsAppHandler.ProxyToWAHA)
		wahaProxy.GET("/:session/screenshot", whatsAppHandler.ProxyToWAHA)

		// Session routes
		wahaProxy.GET("/sessions/:session", whatsAppHandler.ProxyToWAHA)
		wahaProxy.POST("/sessions/:session/restart", whatsAppHandler.ProxyToWAHA)
	}

	// Webhooks
	webhooks := r.Group("/webhooks")
	{
		webhooks.POST("/whatsapp", whatsAppHandler.WebhookHandler)
	}

	log.Printf("[ROUTER] ✅ Todas as rotas configuradas com sucesso!")
	return r
}
