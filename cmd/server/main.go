package main

import (
	"log"
	"os"

	"tappyone/internal/config"
	"tappyone/internal/database"
	"tappyone/internal/router"
	"tappyone/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Carregar variáveis de ambiente
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			if err := godotenv.Load("../../.env"); err != nil {
				log.Println("Arquivo .env não encontrado, usando variáveis do sistema")
			}
		}
	}

	// Configurar aplicação
	cfg := config.Load()

	// Conectar ao banco de dados
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Falha ao conectar com o banco de dados:", err)
	}

	// Executar migrações
	if err := database.Migrate(db); err != nil {
		log.Fatal("Falha ao executar migrações:", err)
	}

	// Conectar ao Redis
	redisClient := database.ConnectRedis(cfg.RedisURL)

	// Inicializar serviços
	serviceContainer := services.NewContainer(db, redisClient, cfg)

	// Configurar modo do Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Configurar rotas
	log.Println("Configurando rotas...")
	r := router.Setup(serviceContainer)
	log.Println("Rotas configuradas com sucesso")

	// Iniciar servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Debug JWT Secret
	jwtSecret := os.Getenv("JWT_SECRET")
	log.Printf("JWT_SECRET configurado: %t (length: %d)", jwtSecret != "", len(jwtSecret))
	if jwtSecret != "" {
		log.Printf("JWT_SECRET preview: %s...", jwtSecret[:10])
	}
	
	log.Printf("Servidor iniciando na porta %s", port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Database URL configurado: %t", cfg.DatabaseURL != "")
	log.Printf("Redis URL configurado: %t", cfg.RedisURL != "")
	
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Falha ao iniciar servidor:", err)
	}
}
