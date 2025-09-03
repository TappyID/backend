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

	log.Println("🗑️ Limpando usuários existentes...")
	
	// PostgreSQL - truncate com cascade
	if err := db.Exec("TRUNCATE TABLE usuarios RESTART IDENTITY CASCADE").Error; err != nil {
		log.Printf("Falha ao truncar, tentando delete: %v", err)
		// Fallback: deletar usuários diretamente
		if err := db.Exec("DELETE FROM usuarios").Error; err != nil {
			log.Printf("Aviso: não foi possível limpar usuários: %v", err)
		}
	}
	
	log.Printf("✅ Usuários existentes removidos")
	
	log.Println("✅ Todos os usuários foram removidos!")
	log.Println("👥 Criando novos usuários...")

	// Criar novos usuários
	if err := createNewUsers(db); err != nil {
		log.Fatal("Erro ao criar usuários:", err)
	}

	log.Println("✅ Usuários criados com sucesso!")
	log.Println("")
	log.Println("🔑 CREDENCIAIS DE ACESSO:")
	log.Println("=====================================")
	log.Println("👨‍💼 ADMIN (Willian): willian@crm.tappy.id / Willian123!")
	log.Println("👨‍💼 ADMIN (Rodrigo): rodrigo@crm.tappy.id / Rodrigo123!")
	log.Println("👤 ATENDENTE: atendente@crm.tappy.id / Atendente123!")
	log.Println("📱 ASSINANTE: assinante@crm.tappy.id / Assinante123!")
	log.Println("=====================================")
}

func createNewUsers(db *gorm.DB) error {
	users := []struct {
		Nome     string
		Email    string
		Password string
		Tipo     models.TipoUsuario
	}{
		{
			Nome:     "Willian Admin",
			Email:    "willian@crm.tappy.id",
			Password: "Willian123!",
			Tipo:     models.TipoUsuarioAdmin,
		},
		{
			Nome:     "Rodrigo Admin",
			Email:    "rodrigo@crm.tappy.id",
			Password: "Rodrigo123!",
			Tipo:     models.TipoUsuarioAdmin,
		},
		{
			Nome:     "Atendente",
			Email:    "atendente@crm.tappy.id",
			Password: "Atendente123!",
			Tipo:     models.TipoUsuarioAtendenteComercial,
		},
		{
			Nome:     "Assinante",
			Email:    "assinante@crm.tappy.id",
			Password: "Assinante123!",
			Tipo:     models.TipoUsuarioAssinante,
		},
	}

	for _, userData := range users {
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

		log.Printf("✅ Usuário criado: %s (%s)", userData.Email, userData.Tipo)
	}

	return nil
}
