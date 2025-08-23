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

	// 1. Criar categoria padrão se não existir
	log.Println("Criando categoria padrão...")
	result := db.Exec(`
		INSERT INTO categorias_respostas (id, nome, descricao, icone, cor, ativo, usuario_id, created_at, updated_at)
		VALUES ('26fc2894-7a13-4719-a0d0-746517c235e8', 'Geral', 'Categoria geral para respostas', 'MessageCircle', '#3b82f6', true, 'ce065849-4fa7-4757-a2cb-5581cfec9225', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`)
	
	if result.Error != nil {
		log.Printf("Erro ao criar categoria: %v", result.Error)
	} else {
		log.Printf("Categoria padrão criada/existe: %d linhas afetadas", result.RowsAffected)
	}

	// 2. Verificar se tabela acoes_respostas existe
	var count int64
	db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'acoes_respostas'").Scan(&count)
	
	if count == 0 {
		log.Println("Tabela acoes_respostas não existe, criando...")
		result = db.Exec(`
			CREATE TABLE acoes_respostas (
				id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
				resposta_rapida_id UUID NOT NULL,
				tipo VARCHAR(50) NOT NULL,
				ordem INTEGER NOT NULL DEFAULT 0,
				delay_segundos INTEGER DEFAULT 0,
				conteudo JSONB NOT NULL,
				obrigatorio BOOLEAN DEFAULT true,
				condicional BOOLEAN DEFAULT false,
				condicao_json TEXT,
				ativo BOOLEAN DEFAULT true,
				created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
				FOREIGN KEY (resposta_rapida_id) REFERENCES resposta_rapidas(id) ON DELETE CASCADE
			)
		`)
		
		if result.Error != nil {
			log.Printf("Erro ao criar tabela acoes_respostas: %v", result.Error)
		} else {
			log.Println("✅ Tabela acoes_respostas criada com sucesso!")
		}
	} else {
		log.Println("✅ Tabela acoes_respostas já existe")
	}

	// 3. Verificar integridade final
	var categoriaCount, respostaCount, acaoCount int64
	db.Raw("SELECT COUNT(*) FROM categorias_respostas").Scan(&categoriaCount)
	db.Raw("SELECT COUNT(*) FROM resposta_rapidas").Scan(&respostaCount)
	db.Raw("SELECT COUNT(*) FROM acoes_respostas").Scan(&acaoCount)
	
	log.Printf("Estado final:")
	log.Printf("- Categorias: %d", categoriaCount)
	log.Printf("- Respostas: %d", respostaCount)
	log.Printf("- Ações: %d", acaoCount)
	log.Println("✅ Banco preparado para criação de respostas!")
}
