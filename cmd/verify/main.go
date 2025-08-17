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
		dsn = "host=localhost user=postgres password=postgres123 dbname=tappyone port=5432 sslmode=disable TimeZone=America/Sao_Paulo"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Falha ao conectar com o banco de dados:", err)
	}

	fmt.Println("=== VERIFICANDO ESTADO ATUAL DA TABELA CARDS ===")

	// Verificar se a tabela existe
	if !db.Migrator().HasTable("cards") {
		fmt.Println("❌ Tabela cards não existe!")
		return
	}
	fmt.Println("✅ Tabela cards existe")

	// Verificar estrutura da coluna conversa_id
	var result struct {
		ColumnName             string  `gorm:"column:column_name"`
		DataType               string  `gorm:"column:data_type"`
		CharacterMaximumLength *int    `gorm:"column:character_maximum_length"`
		IsNullable             string  `gorm:"column:is_nullable"`
	}

	err = db.Raw(`
		SELECT column_name, data_type, character_maximum_length, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'cards' AND column_name = 'conversa_id'
	`).Scan(&result).Error

	if err != nil {
		log.Printf("❌ Erro ao verificar coluna: %v", err)
		return
	}

	fmt.Printf("📋 Coluna conversa_id:\n")
	fmt.Printf("   - Tipo: %s\n", result.DataType)
	if result.CharacterMaximumLength != nil {
		fmt.Printf("   - Tamanho máximo: %d\n", *result.CharacterMaximumLength)
	}
	fmt.Printf("   - Permite NULL: %s\n", result.IsNullable)

	// Verificar constraints
	var constraints []struct {
		ConstraintName string `gorm:"column:constraint_name"`
		ConstraintType string `gorm:"column:constraint_type"`
	}

	err = db.Raw(`
		SELECT constraint_name, constraint_type
		FROM information_schema.table_constraints 
		WHERE table_name = 'cards' AND constraint_name LIKE '%conversa%'
	`).Scan(&constraints).Error

	if err != nil {
		log.Printf("❌ Erro ao verificar constraints: %v", err)
	} else {
		fmt.Printf("🔗 Constraints relacionadas à conversa_id: %d\n", len(constraints))
		for _, c := range constraints {
			fmt.Printf("   - %s (%s)\n", c.ConstraintName, c.ConstraintType)
		}
	}

	// Contar registros na tabela
	var count int64
	db.Table("cards").Count(&count)
	fmt.Printf("📊 Total de cards na tabela: %d\n", count)

	// Tentar inserir um teste
	fmt.Println("\n🧪 TESTANDO INSERÇÃO...")
	testSQL := `
		INSERT INTO cards (id, criado_em, atualizado_em, nome, conversa_id, coluna_id, posicao, ativo)
		VALUES (gen_random_uuid(), NOW(), NOW(), 'Teste', '5518996066756@c.us', gen_random_uuid(), 0, true)
		RETURNING id
	`

	var testID string
	err = db.Raw(testSQL).Scan(&testID).Error
	if err != nil {
		fmt.Printf("❌ Erro ao inserir teste: %v\n", err)
	} else {
		fmt.Printf("✅ Inserção de teste bem-sucedida! ID: %s\n", testID)
		// Limpar o teste
		db.Exec("DELETE FROM cards WHERE id = ?", testID)
		fmt.Println("🧹 Teste limpo")
	}

	fmt.Println("\n=== VERIFICAÇÃO CONCLUÍDA ===")
}
