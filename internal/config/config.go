package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	// Database
	DatabaseURL string

	// Redis
	RedisURL string

	// JWT
	JWTSecret    string
	JWTExpiresIn string

	// WhatsApp API
	WhatsAppAPIURL   string
	WhatsAppAPIToken string
	WebhookURL       string

	// Email SMTP
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	SMTPFrom string

	// DeepSeek AI
	DeepSeekAPIKey string
	DeepSeekAPIURL string

	// Server
	Port        string
	Environment string
}

// loadEnvFile carrega variáveis de um arquivo .env
func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return // Arquivo não existe, ignora
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remover aspas se existirem
			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}
			os.Setenv(key, value)
		}
	}
}

func Load() *Config {
	// Tentar carregar arquivo de configuração
	loadEnvFile("config.env")
	loadEnvFile(".env")

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return &Config{
		// Database
		DatabaseURL: getEnv("DATABASE_URL", "postgresql://postgres:postgres123@localhost:5432/tappyone?sslmode=disable"),

		// Redis
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379"),

		// JWT
		JWTSecret:    getEnv("JWT_SECRET", "sua-chave-secreta-jwt-aqui"),
		JWTExpiresIn: getEnv("JWT_EXPIRES_IN", "7d"),

		// WhatsApp API
		WhatsAppAPIURL:   getEnv("WAHA_API_URL", "https://apiwhatsapp.vyzer.com.br"),
		WhatsAppAPIToken: getEnv("WHATSAPP_API_TOKEN", ""),
		WebhookURL:       getEnv("WEBHOOK_URL", "http://localhost:8081/api/webhooks/whatsapp"),

		// Email SMTP
		SMTPHost: getEnv("SMTP_HOST", "smtp.hostinger.com"),
		SMTPPort: smtpPort,
		SMTPUser: getEnv("SMTP_USER", ""),
		SMTPPass: getEnv("SMTP_PASS", ""),
		SMTPFrom: getEnv("SMTP_FROM", ""),

		// DeepSeek AI
		DeepSeekAPIKey: getEnv("DEEPSEEK_API_KEY", ""),
		DeepSeekAPIURL: getEnv("DEEPSEEK_API_URL", "https://api.deepseek.com"),

		// Server
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("NODE_ENV", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
