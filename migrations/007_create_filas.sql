-- 007_create_filas.sql
-- Migração para criar tabelas de filas e relacionamentos

-- Criar tabela de filas
CREATE TABLE IF NOT EXISTS filas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nome VARCHAR(255) NOT NULL,
    descricao TEXT,
    cor VARCHAR(7) NOT NULL DEFAULT '#3B82F6', -- Hex color
    ordenacao INTEGER NOT NULL DEFAULT 1,
    ativa BOOLEAN NOT NULL DEFAULT true,
    prioridade VARCHAR(20) NOT NULL DEFAULT 'MEDIA' CHECK (prioridade IN ('BAIXA', 'MEDIA', 'ALTA', 'URGENTE')),
    chat_bot BOOLEAN NOT NULL DEFAULT false,
    kanban BOOLEAN NOT NULL DEFAULT false,
    whatsapp_chats BOOLEAN NOT NULL DEFAULT false,
    criado_em TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    atualizado_em TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Criar tabela de relacionamento many-to-many entre filas e atendentes
CREATE TABLE IF NOT EXISTS fila_atendentes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fila_id UUID NOT NULL REFERENCES filas(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    criado_em TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(fila_id, user_id) -- Evitar duplicatas
);

-- Criar índices para performance
CREATE INDEX IF NOT EXISTS idx_filas_ordenacao ON filas(ordenacao);
CREATE INDEX IF NOT EXISTS idx_filas_ativa ON filas(ativa);
CREATE INDEX IF NOT EXISTS idx_filas_prioridade ON filas(prioridade);
CREATE INDEX IF NOT EXISTS idx_fila_atendentes_fila_id ON fila_atendentes(fila_id);
CREATE INDEX IF NOT EXISTS idx_fila_atendentes_user_id ON fila_atendentes(user_id);

-- Trigger para atualizar automaticamente o campo atualizado_em
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.atualizado_em = NOW();
   RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_filas_updated_at 
    BEFORE UPDATE ON filas 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Inserir dados de exemplo (opcional)
INSERT INTO filas (nome, descricao, cor, ordenacao, ativa, prioridade, chat_bot, kanban, whatsapp_chats) VALUES
('Suporte', 'Fila para atendimento de suporte técnico', '#EF4444', 1, true, 'ALTA', true, true, true),
('Vendas', 'Fila para atendimento de vendas', '#10B981', 2, true, 'MEDIA', true, true, true),
('Cobrança', 'Fila para cobrança e financeiro', '#F59E0B', 3, true, 'MEDIA', false, true, true)
ON CONFLICT DO NOTHING;
