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

	fmt.Println("=== FORÇANDO MIGRAÇÃO DA TABELA CARDS ===")

	// 1. Dropar tabelas relacionadas primeiro
	fmt.Println("🗑️ Dropando tabelas relacionadas...")
	db.Exec("DROP TABLE IF EXISTS card_resposta_rapidas CASCADE")
	
	// 2. Dropar tabela cards
	fmt.Println("🗑️ Dropando tabela cards...")
	db.Exec("DROP TABLE IF EXISTS cards CASCADE")
	
	// 3. Recriar tabela cards com estrutura correta
	fmt.Println("🏗️ Recriando tabela cards...")
	createSQL := `
		CREATE TABLE cards (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			criado_em TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			atualizado_em TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			nome VARCHAR(255) NOT NULL,
			descricao TEXT,
			posicao INTEGER NOT NULL DEFAULT 0,
			coluna_id UUID NOT NULL,
			conversa_id VARCHAR(255) NOT NULL,
			prioridade INTEGER DEFAULT 0,
			data_vencimento TIMESTAMP WITH TIME ZONE,
			ativo BOOLEAN DEFAULT true
		)
	`
	
	err = db.Exec(createSQL).Error
	if err != nil {
		log.Printf("❌ Erro ao criar tabela: %v", err)
		os.Exit(1)
	}
	
	// 4. Criar índices
	fmt.Println("📊 Criando índices...")
	db.Exec("CREATE INDEX idx_cards_coluna_id ON cards(coluna_id)")
	db.Exec("CREATE INDEX idx_cards_conversa_id ON cards(conversa_id)")
	db.Exec("CREATE INDEX idx_cards_ativo ON cards(ativo)")
	
	// 5. Verificar resultado
	fmt.Println("✅ Verificando resultado...")
	var result struct {
		ColumnName             string  `gorm:"column:column_name"`
		DataType               string  `gorm:"column:data_type"`
		CharacterMaximumLength *int    `gorm:"column:character_maximum_length"`
	}

	err = db.Raw(`
		SELECT column_name, data_type, character_maximum_length 
		FROM information_schema.columns 
		WHERE table_name = 'cards' AND column_name = 'conversa_id'
	`).Scan(&result).Error

	if err != nil {
		log.Printf("❌ Erro ao verificar resultado: %v", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Coluna conversa_id criada com sucesso!\n")
	fmt.Printf("   - Tipo: %s\n", result.DataType)
	if result.CharacterMaximumLength != nil {
		fmt.Printf("   - Tamanho: %d\n", *result.CharacterMaximumLength)
	}

	// 6. Teste de inserção
	fmt.Println("🧪 Testando inserção...")
	testSQL := `
		INSERT INTO cards (nome, conversa_id, coluna_id, posicao, ativo)
		VALUES ('Teste WhatsApp', '5518996066756@c.us', gen_random_uuid(), 0, true)
		RETURNING id
	`

	var testID string
	err = db.Raw(testSQL).Scan(&testID).Error
	if err != nil {
		log.Printf("❌ Erro no teste: %v", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Teste bem-sucedido! ID: %s\n", testID)
	
	// Limpar teste
	db.Exec("DELETE FROM cards WHERE id = ?", testID)
	
	fmt.Println("🎉 MIGRAÇÃO FORÇADA CONCLUÍDA COM SUCESSO!")
	fmt.Println("📝 Agora você pode reiniciar o backend e testar o Kanban.")
}
