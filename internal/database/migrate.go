package database

import (
	"log"
	"tappyone/internal/models"

	"gorm.io/gorm"
)

// Migrate executa as migrações do banco de dados
func Migrate(db *gorm.DB) error {
	log.Printf("[MIGRATION] Starting database migration...")
	
	// Migração manual para corrigir tipo da coluna conversa_id
	log.Printf("[MIGRATION] Executing fixConversaIdColumnType...")
	if err := fixConversaIdColumnType(db); err != nil {
		log.Printf("[MIGRATION] Error in fixConversaIdColumnType: %v", err)
		return err
	}
	
	log.Printf("[MIGRATION] Running AutoMigrate...")
	err := db.AutoMigrate(
		// Usuários e autenticação
		&models.Usuario{},
		
		// WhatsApp
		&models.SessaoWhatsApp{},
		&models.Contato{},
		&models.Conversa{},
		&models.Mensagem{},
		
		// Tags
		&models.Tag{},
		&models.ContatoTag{},
		
		// Kanban
		&models.Quadro{},
		&models.QuadroTag{},
		&models.Coluna{},
		// &models.Card{}, // DESABILITADO TEMPORARIAMENTE
		
		// Respostas rápidas
		&models.RespostaRapida{},
		// &models.CardRespostaRapida{}, // DESABILITADO TEMPORARIAMENTE
		
		// IA
		&models.AgenteIa{},
		
		// Atendimento
		&models.Atendimento{},
		&models.Agendamento{},
		
		// Chat interno
		&models.MensagemInterna{},
		
		// NPS
		&models.AvaliacaoNps{},
		
		// Fluxos
		&models.Fluxo{},
		&models.FluxoNo{},
		&models.FluxoConexao{},
		
		// Planos e cobrança
		&models.Plano{},
		&models.Assinatura{},
		&models.Cobranca{},
	)
	
	// Forçar criação manual da tabela cards com estrutura correta
	log.Printf("[MIGRATION] Creating cards table manually...")
	err2 := createCardsTableManually(db)
	if err2 != nil {
		log.Printf("[MIGRATION] Error creating cards table: %v", err2)
		return err2
	}
	
	if err != nil {
		log.Printf("[MIGRATION] AutoMigrate error: %v", err)
		return err
	}
	
	log.Printf("[MIGRATION] Migration completed successfully")
	return nil
}

// fixConversaIdColumnType corrige o tipo da coluna conversa_id na tabela cards
func fixConversaIdColumnType(db *gorm.DB) error {
	log.Printf("[MIGRATION] Starting fixConversaIdColumnType...")
	
	// Verificar se a tabela cards existe
	if !db.Migrator().HasTable("cards") {
		log.Printf("[MIGRATION] Table cards does not exist, will be created correctly")
		return nil
	}
	
	log.Printf("[MIGRATION] Table cards exists, checking column type...")
	
	// Executar SQL direto para corrigir a coluna
	log.Printf("[MIGRATION] Executing direct SQL to fix conversa_id column...")
	
	// 1. Remover constraints
	db.Exec("ALTER TABLE cards DROP CONSTRAINT IF EXISTS fk_conversas_cards")
	db.Exec("ALTER TABLE cards DROP CONSTRAINT IF EXISTS fk_cards_conversa")
	db.Exec("ALTER TABLE cards DROP CONSTRAINT IF EXISTS cards_conversa_id_fkey")
	log.Printf("[MIGRATION] Dropped foreign key constraints")
	
	// 2. Limpar dados existentes
	result := db.Exec("DELETE FROM cards WHERE conversa_id IS NOT NULL")
	log.Printf("[MIGRATION] Deleted %d existing cards", result.RowsAffected)
	
	// 3. Alterar tipo da coluna
	err := db.Exec("ALTER TABLE cards ALTER COLUMN conversa_id TYPE VARCHAR(255) USING conversa_id::VARCHAR(255)").Error
	if err != nil {
		log.Printf("[MIGRATION] Error altering column type: %v", err)
		return err
	}
	
	log.Printf("[MIGRATION] Successfully altered conversa_id column to VARCHAR(255)")
	return nil
}

// createCardsTableManually cria a tabela cards com a estrutura correta
func createCardsTableManually(db *gorm.DB) error {
	log.Printf("[MIGRATION] Checking if cards table needs manual creation...")
	
	// Dropar tabela se existir
	log.Printf("[MIGRATION] Dropping existing cards table...")
	db.Exec("DROP TABLE IF EXISTS card_resposta_rapidas CASCADE")
	db.Exec("DROP TABLE IF EXISTS cards CASCADE")
	
	// Criar tabela com estrutura correta
	log.Printf("[MIGRATION] Creating cards table with correct structure...")
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
	
	err := db.Exec(createSQL).Error
	if err != nil {
		log.Printf("[MIGRATION] Error creating cards table: %v", err)
		return err
	}
	
	// Criar índices
	log.Printf("[MIGRATION] Creating indexes...")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_cards_coluna_id ON cards(coluna_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_cards_conversa_id ON cards(conversa_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_cards_ativo ON cards(ativo)")
	
	log.Printf("[MIGRATION] Cards table created successfully with VARCHAR conversa_id")
	return nil
}
