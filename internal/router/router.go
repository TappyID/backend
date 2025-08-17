package router

import (
	"fmt"
	"log"
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

		// WhatsApp API (sem middleware JWT, usa autenticação manual)
		whatsappAPI := r.Group("/api/whatsapp")
		{
			whatsappAPI.GET("/chats", func(c *gin.Context) {
				log.Printf("[WHATSAPP] GET /chats - Starting request")
				
				// Autenticação via header Authorization
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					log.Printf("[WHATSAPP] GET /chats - Auth failed: %v", err)
					c.JSON(401, gin.H{"error": "Invalid token", "details": err.Error()})
					return
				}
				
				sessionName := fmt.Sprintf("user_%s", userID)
				log.Printf("[WHATSAPP] GET /chats - UserID: %s, SessionName: %s", userID, sessionName)
				
				var chats interface{}
				chats, err = container.WhatsAppService.GetChats(sessionName)
				if err != nil {
					log.Printf("[WHATSAPP] GET /chats - WhatsAppService.GetChats failed: %v", err)
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				
				log.Printf("[WHATSAPP] GET /chats - Success, returning chats")
				c.JSON(200, chats)
			})
			
			whatsappAPI.GET("/contacts", func(c *gin.Context) {
				// Autenticação via header Authorization
				userID, err := utils.ValidateJWTFromHeader(c, container.AuthService)
				if err != nil {
					c.JSON(401, gin.H{"error": "Invalid token", "details": err.Error()})
					return
				}
				sessionName := fmt.Sprintf("user_%s", userID)
				
				var contacts interface{}
				contacts, err = container.WhatsAppService.GetContacts(sessionName)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, contacts)
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
				messages, err = container.WhatsAppService.GetChatMessages(sessionName, chatID)
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
					Text string `json:"text" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": "Invalid request"})
					return
				}
				
				var result interface{}
				result, err = container.WhatsAppService.SendMessage(sessionName, chatID, req.Text)
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
			
			// Endpoint para obter informações de presença
			whatsappAPI.GET("/presence", func(c *gin.Context) {
				userID, exists := c.Get("userID")
				if !exists {
					c.JSON(401, gin.H{"error": "User not authenticated"})
					return
				}
				
				sessionName := fmt.Sprintf("user_%s", userID)
				result, err := container.WhatsAppService.GetPresence(sessionName)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})
			
			// Endpoint para definir status de digitação
			whatsappAPI.POST("/chats/:chatId/typing", func(c *gin.Context) {
				userID, exists := c.Get("userID")
				if !exists {
					c.JSON(401, gin.H{"error": "User not authenticated"})
					return
				}
				
				sessionName := fmt.Sprintf("user_%s", userID)
				chatID := c.Param("chatId")
				
				var req struct {
					Presence string `json:"presence" binding:"required"` // "typing" ou "paused"
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(400, gin.H{"error": "Invalid request"})
					return
				}
				
				result, err := container.WhatsAppService.SetTyping(sessionName, chatID, req.Presence)
				if err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
					return
				}
				c.JSON(200, result)
			})
			
			// Rotas de mídia
			whatsappAPI.POST("/upload", whatsappMediaHandler.UploadFile)
			whatsappAPI.POST("/chats/:chatId/media", whatsappMediaHandler.UploadAndSendMedia)
			whatsappAPI.POST("/chats/:chatId/image", whatsappMediaHandler.SendImage)
			whatsappAPI.POST("/chats/:chatId/file", whatsappMediaHandler.SendFile)
			whatsappAPI.POST("/chats/:chatId/voice", whatsappMediaHandler.SendVoice)
			whatsappAPI.POST("/chats/:chatId/video", whatsappMediaHandler.SendVideo)
		}
	}



	// Webhooks
	webhooks := r.Group("/webhooks")
	{
		webhooks.POST("/whatsapp", whatsappHandler.WebhookHandler)
	}

	return r
}
