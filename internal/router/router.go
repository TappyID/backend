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

	// Configurar CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:3001"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// Middleware de logging
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(container.AuthService)
	userHandler := handlers.NewUserHandler(container.UserService)
	whatsappHandler := handlers.NewWhatsAppHandler(container.WhatsAppService)
	whatsappMediaHandler := handlers.NewWhatsAppMediaHandler(container.WhatsAppService, container.AuthService)
	kanbanHandler := handlers.NewKanbanHandler(container.KanbanService)
	connectionHandler := handlers.NewConnectionHandler(container.ConnectionService)
	respostaRapidaHandler := handlers.NewRespostaRapidaHandler(container.RespostaRapidaService)

	// Servir arquivos estáticos (uploads)
	r.Static("/uploads", "./uploads")

	// Rotas públicas
	public := r.Group("/api")
	{
		public.POST("/auth/login", authHandler.Login)
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "message": "TappyOne CRM API"})
		})
	}

	// WebSocket route (with JWT auth via query param)
	r.GET("/ws", handlers.NewWebSocketHandler(container.AuthService, container.Config.JWTSecret))

	// Webhook para receber mensagens da WAHA API
	r.POST("/webhook/whatsapp", func(c *gin.Context) {
		log.Printf("[WEBHOOK] Received WhatsApp webhook")

		var payload map[string]interface{}
		if err := c.ShouldBindJSON(&payload); err != nil {
			log.Printf("[WEBHOOK] Error parsing payload: %v", err)
			c.JSON(400, gin.H{"error": "Invalid payload"})
			return
		}

		log.Printf("[WEBHOOK] Payload: %+v", payload)

		// Verificar se é uma mensagem nova
		if event, ok := payload["event"].(string); ok && event == "message" {
			if data, ok := payload["data"].(map[string]interface{}); ok {
				// Extrair session da mensagem para identificar o usuário
				if session, ok := payload["session"].(string); ok {
					// Extrair userID do session name (formato: user_UUID)
					if len(session) > 5 && session[:5] == "user_" {
						userID := session[5:]
						log.Printf("[WEBHOOK] Broadcasting message to user: %s", userID)

						// Enviar via WebSocket
						handlers.BroadcastNewMessage(userID, data)
					}
				}
			}
		}

		c.JSON(200, gin.H{"status": "ok"})
	})

	// Webhook específico para respostas rápidas
	r.POST("/webhook/resposta-rapida", respostaRapidaHandler.ProcessarMensagemWebhook)

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
			whatsapp.POST("/sessions", whatsappHandler.CreateSession)
			whatsapp.GET("/sessions", whatsappHandler.ListSessions)
			whatsapp.GET("/sessions/:id", whatsappHandler.GetSession)
			whatsapp.POST("/sessions/:id/start", whatsappHandler.StartSession)
			whatsapp.POST("/sessions/:id/stop", whatsappHandler.StopSession)
			whatsapp.GET("/sessions/:id/qr", whatsappHandler.GetQRCode)
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
			connections.GET("/", connectionHandler.GetUserConnections)
			connections.GET("/whatsapp", connectionHandler.GetUserConnection)
			connections.GET("/:platform", connectionHandler.GetUserConnection)
			connections.POST("/", connectionHandler.CreateOrUpdateConnection)
			connections.PUT("/:platform", connectionHandler.UpdateConnection)
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

			// Respostas Rápidas (rotas com :id por último)
			respostasRapidas.GET("/", respostaRapidaHandler.GetRespostasRapidas)
			respostasRapidas.POST("/", respostaRapidaHandler.CreateRespostaRapida)
			respostasRapidas.GET("/:id", respostaRapidaHandler.GetRespostaRapida)
			respostasRapidas.PUT("/:id/pausar", respostaRapidaHandler.TogglePausarRespostaRapida)
			respostasRapidas.POST("/:id/executar", respostaRapidaHandler.ExecutarRespostaRapida)
			respostasRapidas.GET("/:id/acoes", respostaRapidaHandler.GetAcoes)
			respostasRapidas.POST("/:id/acoes", respostaRapidaHandler.CreateAcao)
			respostasRapidas.PUT("/:id/acoes/reorder", respostaRapidaHandler.ReorderAcoes)
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

			// Rotas de mídia
			whatsappAPI.POST("/chats/:chatId/voice", whatsappHandler.SendVoiceMessage)
			whatsappAPI.POST("/chats/:chatId/image", whatsappHandler.SendImageMessage)
			whatsappAPI.POST("/chats/:chatId/video", whatsappMediaHandler.SendVideoMessage)
			whatsappAPI.POST("/chats/:chatId/file", whatsappHandler.SendFileMessage)
			whatsappAPI.GET("/media/:mediaId", whatsappHandler.DownloadMedia)
			
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

	// Webhooks
	webhooks := r.Group("/webhooks")
	{
		webhooks.POST("/whatsapp", whatsappHandler.WebhookHandler)
	}

	return r
}
