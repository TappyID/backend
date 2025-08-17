package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Configurar conexão com o banco
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Usar as mesmas configurações do projeto
		dsn = "host=localhost user=postgres password=postgres123 dbname=tappyone port=5432 sslmode=disable TimeZone=America/Sao_Paulo"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Falha ao conectar com o banco de dados:", err)
	}

	fmt.Println("=== EXECUTANDO MIGRAÇÃO MANUAL ===")

	// Verificar tipo atual da coluna
	var result struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
	}

	err = db.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'cards' AND column_name = 'conversa_id'").Scan(&result).Error
	if err != nil {
		log.Printf("Erro ao verificar coluna: %v", err)
	} else {
		fmt.Printf("Tipo atual da coluna conversa_id: %s\n", result.DataType)
	}

	// Executar migração
	fmt.Println("Removendo constraints...")
	db.Exec("ALTER TABLE cards DROP CONSTRAINT IF EXISTS fk_conversas_cards")
	db.Exec("ALTER TABLE cards DROP CONSTRAINT IF EXISTS fk_cards_conversa")
	db.Exec("ALTER TABLE cards DROP CONSTRAINT IF EXISTS cards_conversa_id_fkey")

	fmt.Println("Limpando dados existentes...")
	result2 := db.Exec("DELETE FROM cards WHERE conversa_id IS NOT NULL")
	fmt.Printf("Deletados %d cards existentes\n", result2.RowsAffected)

	fmt.Println("Alterando tipo da coluna...")
	err = db.Exec("ALTER TABLE cards ALTER COLUMN conversa_id TYPE VARCHAR(255) USING conversa_id::VARCHAR(255)").Error
	if err != nil {
		log.Printf("Erro ao alterar coluna: %v", err)
		os.Exit(1)
	}

	// Verificar resultado
	err = db.Raw("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'cards' AND column_name = 'conversa_id'").Scan(&result).Error
	if err != nil {
		log.Printf("Erro ao verificar resultado: %v", err)
	} else {
		fmt.Printf("Novo tipo da coluna conversa_id: %s\n", result.DataType)
	}

	fmt.Println("=== MIGRAÇÃO CONCLUÍDA COM SUCESSO ===")
}
