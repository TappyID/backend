package services

import (
	"tappyone/internal/config"
	"tappyone/internal/repositories"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Container contém todos os serviços da aplicação
type Container struct {
	DB     *gorm.DB
	Redis  *redis.Client
	Config *config.Config

	// Serviços
	AuthService         *AuthService
	UserService         *UserService
	WhatsAppService     *WhatsAppService
	KanbanService       *KanbanService
	MessageService      *MessageService
	AIService           *AIService
	EmailService        *EmailService
	ConnectionService   *ConnectionService
	RespostaRapidaService *RespostaRapidaService
}

// NewContainer cria uma nova instância do container de serviços
func NewContainer(db *gorm.DB, redis *redis.Client, cfg *config.Config) *Container {
	container := &Container{
		DB:     db,
		Redis:  redis,
		Config: cfg,
	}

	// Inicializar serviços
	container.AuthService = NewAuthService(db, redis, cfg)
	container.UserService = NewUserService(db)
	container.WhatsAppService = NewWhatsAppService(db, cfg)
	container.KanbanService = NewKanbanService(db)
	container.MessageService = NewMessageService(db, redis)
	container.AIService = NewAIService(cfg)
	container.EmailService = NewEmailService(cfg)
	
	// Inicializar repositórios e serviços de conexão
	connectionRepo := repositories.NewConnectionRepository(db)
	container.ConnectionService = NewConnectionService(connectionRepo, "https://apiwhatsapp.vyzer.com.br/api", "atendia-waha-2024-secretkey")
	
	// Inicializar serviço de respostas rápidas
	respostaRapidaRepo := repositories.NewRespostaRapidaRepository(db)
	container.RespostaRapidaService = NewRespostaRapidaService(respostaRapidaRepo, container.WhatsAppService)

	return container
}
