-- Migration 006: Adicionar modelos dos 4 fluxos (Agendamento, Orçamento, Assinatura, Anotação)
-- Data: 2025-08-22

-- 1. Adicionar campo linkMeeting na tabela agendamentos
ALTER TABLE agendamentos ADD COLUMN link_meeting TEXT;

-- 2. Atualizar tabela assinaturas com novos campos
ALTER TABLE assinaturas ADD COLUMN nome VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE assinaturas ADD COLUMN contato_id VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE assinaturas ADD COLUMN forma_pagamento VARCHAR(50) NOT NULL DEFAULT 'PIX';
ALTER TABLE assinaturas ADD COLUMN link_pagamento TEXT;
ALTER TABLE assinaturas ADD COLUMN valor DECIMAL(10,2) NOT NULL DEFAULT 0.00;
ALTER TABLE assinaturas ADD COLUMN renovacao VARCHAR(50) NOT NULL DEFAULT 'mensal';

-- Adicionar foreign key para contato
ALTER TABLE assinaturas ADD CONSTRAINT fk_assinaturas_contato 
    FOREIGN KEY (contato_id) REFERENCES contatos(id) ON DELETE CASCADE;

-- 3. Criar tabela orcamentos
CREATE TABLE orcamentos (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    titulo VARCHAR(255) NOT NULL,
    data TIMESTAMP NOT NULL,
    tipo VARCHAR(50) NOT NULL DEFAULT 'orcamento',
    observacao TEXT,
    valor_total DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDENTE',
    usuario_id VARCHAR(255) NOT NULL,
    contato_id VARCHAR(255) NOT NULL,
    criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    atualizado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_orcamentos_usuario FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_orcamentos_contato FOREIGN KEY (contato_id) REFERENCES contatos(id) ON DELETE CASCADE
);

-- Índices para orcamentos
CREATE INDEX idx_orcamentos_usuario_id ON orcamentos(usuario_id);
CREATE INDEX idx_orcamentos_contato_id ON orcamentos(contato_id);
CREATE INDEX idx_orcamentos_status ON orcamentos(status);
CREATE INDEX idx_orcamentos_data ON orcamentos(data);

-- 4. Criar tabela orcamento_itens
CREATE TABLE orcamento_itens (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    nome VARCHAR(255) NOT NULL,
    valor DECIMAL(10,2) NOT NULL,
    quantidade INTEGER NOT NULL DEFAULT 1,
    subtotal DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    orcamento_id VARCHAR(255) NOT NULL,
    criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    atualizado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_orcamento_itens_orcamento FOREIGN KEY (orcamento_id) REFERENCES orcamentos(id) ON DELETE CASCADE
);

-- Índices para orcamento_itens
CREATE INDEX idx_orcamento_itens_orcamento_id ON orcamento_itens(orcamento_id);

-- 5. Criar tabela anotacoes
CREATE TABLE anotacoes (
    id VARCHAR(255) PRIMARY KEY DEFAULT gen_random_uuid(),
    titulo VARCHAR(255) NOT NULL,
    conteudo TEXT NOT NULL,
    importante BOOLEAN NOT NULL DEFAULT FALSE,
    usuario_id VARCHAR(255) NOT NULL,
    contato_id VARCHAR(255) NOT NULL,
    criado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    atualizado_em TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_anotacoes_usuario FOREIGN KEY (usuario_id) REFERENCES usuarios(id) ON DELETE CASCADE,
    CONSTRAINT fk_anotacoes_contato FOREIGN KEY (contato_id) REFERENCES contatos(id) ON DELETE CASCADE
);

-- Índices para anotacoes
CREATE INDEX idx_anotacoes_usuario_id ON anotacoes(usuario_id);
CREATE INDEX idx_anotacoes_contato_id ON anotacoes(contato_id);
CREATE INDEX idx_anotacoes_importante ON anotacoes(importante);

-- Triggers para atualizar automaticamente atualizado_em
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.atualizado_em = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Aplicar triggers nas novas tabelas
CREATE TRIGGER update_orcamentos_updated_at 
    BEFORE UPDATE ON orcamentos 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_orcamento_itens_updated_at 
    BEFORE UPDATE ON orcamento_itens 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_anotacoes_updated_at 
    BEFORE UPDATE ON anotacoes 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comentários para documentação
COMMENT ON TABLE orcamentos IS 'Tabela para armazenar orçamentos dos usuários';
COMMENT ON TABLE orcamento_itens IS 'Tabela para armazenar itens dos orçamentos';
COMMENT ON TABLE anotacoes IS 'Tabela para armazenar anotações dos contatos pelos usuários';

COMMENT ON COLUMN orcamentos.tipo IS 'Tipo do orçamento: venda, assinatura, orcamento, cobranca';
COMMENT ON COLUMN orcamentos.status IS 'Status do orçamento: PENDENTE, APROVADO, REJEITADO, CANCELADO';
COMMENT ON COLUMN assinaturas.forma_pagamento IS 'Forma de pagamento: PIX, CARTAO, BOLETO';
COMMENT ON COLUMN assinaturas.renovacao IS 'Período de renovação: mensal, trimestral, semestral, anual, limitado';
