package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Carregar variáveis de ambiente
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("Arquivo .env não encontrado, usando variáveis do sistema")
	}

	// Conectar ao banco
	db, err := gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		log.Fatal("Falha ao conectar com o banco:", err)
	}

	log.Println("Conectado ao banco PostgreSQL")

	// Limpar dados inconsistentes
	log.Println("Limpando dados inconsistentes...")

	// 1. Deletar respostas rápidas órfãs (sem categoria válida)
	result := db.Exec(`
		DELETE FROM resposta_rapidas 
		WHERE categoria_id NOT IN (SELECT id FROM categorias_respostas)
	`)
	
	if result.Error != nil {
		log.Printf("Erro ao deletar respostas órfãs: %v", result.Error)
	} else {
		log.Printf("Deletadas %d respostas rápidas órfãs", result.RowsAffected)
	}

	// 2. Deletar ações órfãs (sem resposta válida)
	result = db.Exec(`
		DELETE FROM acoes_respostas 
		WHERE resposta_rapida_id NOT IN (SELECT id FROM resposta_rapidas)
	`)
	
	if result.Error != nil {
		log.Printf("Erro ao deletar ações órfãs: %v", result.Error)
	} else {
		log.Printf("Deletadas %d ações órfãs", result.RowsAffected)
	}

	// 3. Verificar integridade
	var count int64
	db.Raw(`
		SELECT COUNT(*) FROM resposta_rapidas r 
		LEFT JOIN categorias_respostas c ON r.categoria_id = c.id 
		WHERE c.id IS NULL
	`).Scan(&count)

	log.Printf("Respostas órfãs restantes: %d", count)

	if count == 0 {
		log.Println("✅ Dados limpos com sucesso!")
	} else {
		log.Println("❌ Ainda existem dados inconsistentes")
	}
}
