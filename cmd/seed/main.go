package main

import (
	"log"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"tappyone/internal/config"
	"tappyone/internal/database"
	"tappyone/internal/models"
)

func main() {
	// Carregar variáveis de ambiente
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis do sistema")
	}

	// Carregar configuração
	cfg := config.Load()

	// Conectar ao banco de dados
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Falha ao conectar com o banco de dados:", err)
	}

	log.Println("Criando usuários de teste...")

	// Criar usuários de teste
	if err := createTestUsers(db); err != nil {
		log.Fatal("Erro ao criar usuários de teste:", err)
	}

	log.Println("Usuários de teste criados com sucesso!")
	log.Println("Credenciais de teste:")
	log.Println("ADMIN: admin@tappyone.com / admin123")
	log.Println("FUNCIONARIO: funcionario@tappyone.com / func123")
	log.Println("ASSINANTE: assinante@tappyone.com / assin123")
}

func createTestUsers(db *gorm.DB) error {
	users := []struct {
		Nome     string
		Email    string
		Password string
		Tipo     models.TipoUsuario
	}{
		{
			Nome:     "Administrador",
			Email:    "admin@tappyone.com",
			Password: "admin123",
			Tipo:     models.TipoUsuarioAdmin,
		},
		{
			Nome:     "Funcionário Teste",
			Email:    "funcionario@tappyone.com",
			Password: "func123",
			Tipo:     models.TipoUsuarioAtendenteComercial,
		},
		{
			Nome:     "Assinante Teste",
			Email:    "assinante@tappyone.com",
			Password: "assin123",
			Tipo:     models.TipoUsuarioAssinante,
		},
	}

	for _, userData := range users {
		// Verificar se usuário já existe
		var existingUser models.Usuario
		result := db.Where("email = ?", userData.Email).First(&existingUser)
		if result.Error == nil {
			log.Printf("Usuário %s já existe, pulando...", userData.Email)
			continue
		}

		// Hash da senha
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		// Criar usuário
		telefone := "+55 11 99999-9999"
		user := models.Usuario{
			Nome:     userData.Nome,
			Email:    userData.Email,
			Senha:    string(hashedPassword),
			Tipo:     userData.Tipo,
			Ativo:    true,
			Telefone: &telefone,
		}

		if err := db.Create(&user).Error; err != nil {
			return err
		}

		log.Printf("Usuário criado: %s (%s)", userData.Email, userData.Tipo)
	}

	return nil
}
