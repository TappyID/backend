-- Migration: Create Respostas Rápidas System
-- Criado em: 2025-01-04

-- Tabela de Categorias de Respostas Rápidas
CREATE TABLE IF NOT EXISTS categorias_respostas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nome VARCHAR(100) NOT NULL,
    descricao TEXT,
    cor VARCHAR(7) NOT NULL DEFAULT '#3B82F6', -- Hex color
    icone VARCHAR(50) DEFAULT 'Zap',
    usuario_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    ativo BOOLEAN DEFAULT true,
    ordem INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de Respostas Rápidas
CREATE TABLE IF NOT EXISTS respostas_rapidas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    titulo VARCHAR(200) NOT NULL,
    descricao TEXT,
    categoria_id UUID NOT NULL REFERENCES categorias_respostas(id) ON DELETE CASCADE,
    usuario_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    
    -- Configurações de agendamento
    agendamento_ativo BOOLEAN DEFAULT false,
    trigger_tipo VARCHAR(50) DEFAULT 'manual', -- 'manual', 'primeira_mensagem', 'palavra_chave', 'horario'
    trigger_condicao TEXT, -- JSON com condições específicas
    
    -- Configurações de execução
    delay_segundos INTEGER DEFAULT 0, -- Delay antes de executar
    repetir BOOLEAN DEFAULT false,
    intervalo_repeticao INTEGER, -- Em minutos
    max_repeticoes INTEGER DEFAULT 1,
    
    -- Configurações de contato
    aplicar_novos_contatos BOOLEAN DEFAULT true,
    aplicar_contatos_existentes BOOLEAN DEFAULT false,
    contatos_especificos TEXT, -- JSON array com IDs específicos
    
    -- Status
    ativo BOOLEAN DEFAULT true,
    pausado BOOLEAN DEFAULT false,
    ordem INTEGER DEFAULT 0,
    
    -- Estatísticas
    total_execucoes INTEGER DEFAULT 0,
    ultima_execucao TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de Ações das Respostas Rápidas
CREATE TABLE IF NOT EXISTS acoes_respostas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resposta_rapida_id UUID NOT NULL REFERENCES respostas_rapidas(id) ON DELETE CASCADE,
    
    -- Tipo e configuração da ação
    tipo VARCHAR(50) NOT NULL, -- 'texto', 'audio', 'imagem', 'video', 'arquivo', 'pix', 'link', 'contato', 'localizacao'
    ordem INTEGER NOT NULL DEFAULT 0,
    delay_segundos INTEGER DEFAULT 0, -- Delay antes desta ação específica
    
    -- Conteúdo da ação (JSON flexível)
    conteudo JSONB NOT NULL,
    /*
    Exemplos de conteúdo por tipo:
    
    texto: {
        "mensagem": "Olá! Como posso ajudar?",
        "usar_ia": false,
        "modelo_ia": "gpt-3.5-turbo",
        "variaveis": ["nome_cliente", "horario_atual"]
    }
    
    audio: {
        "tipo": "arquivo", // 'arquivo' ou 'ia_gerado'
        "arquivo_url": "/uploads/audio.mp3",
        "texto_ia": "Olá, tudo bem?",
        "voz_ia": "pt-BR-female"
    }
    
    pix: {
        "tipo_chave": "cpf", // 'cpf', 'email', 'telefone', 'aleatoria'
        "chave": "123.456.789-00",
        "valor": 50.00,
        "descricao": "Pagamento do serviço"
    }
    
    imagem: {
        "arquivo_url": "/uploads/imagem.jpg",
        "legenda": "Nossos produtos"
    }
    */
    
    -- Configurações específicas
    obrigatorio BOOLEAN DEFAULT true, -- Se falhar, para a sequência?
    condicional BOOLEAN DEFAULT false,
    condicao_json TEXT, -- Condições para executar esta ação
    
    -- Status
    ativo BOOLEAN DEFAULT true,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de Execuções de Respostas Rápidas (Log/Histórico)
CREATE TABLE IF NOT EXISTS execucoes_respostas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resposta_rapida_id UUID NOT NULL REFERENCES respostas_rapidas(id) ON DELETE CASCADE,
    usuario_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    
    -- Informações do contato/chat
    chat_id VARCHAR(255) NOT NULL,
    contato_nome VARCHAR(255),
    contato_telefone VARCHAR(50),
    
    -- Informações da execução
    trigger_tipo VARCHAR(50) NOT NULL,
    trigger_dados JSONB, -- Dados que dispararam a execução
    
    -- Status da execução
    status VARCHAR(50) DEFAULT 'pendente', -- 'pendente', 'executando', 'concluida', 'erro', 'cancelada'
    acoes_executadas INTEGER DEFAULT 0,
    total_acoes INTEGER DEFAULT 0,
    erro_mensagem TEXT,
    
    -- Agendamento
    agendado_para TIMESTAMP,
    iniciado_em TIMESTAMP,
    concluido_em TIMESTAMP,
    
    -- Resultados
    mensagens_enviadas INTEGER DEFAULT 0,
    resultado_json JSONB, -- Detalhes dos resultados de cada ação
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de Agendamentos Ativos (Para controle em background)
CREATE TABLE IF NOT EXISTS agendamentos_respostas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    resposta_rapida_id UUID NOT NULL REFERENCES respostas_rapidas(id) ON DELETE CASCADE,
    usuario_id UUID NOT NULL REFERENCES usuarios(id) ON DELETE CASCADE,
    
    -- Informações do contato
    chat_id VARCHAR(255) NOT NULL,
    contato_nome VARCHAR(255),
    contato_telefone VARCHAR(50),
    
    -- Configurações do agendamento
    trigger_tipo VARCHAR(50) NOT NULL,
    condicoes_json JSONB,
    
    -- Controle de execução
    ativo BOOLEAN DEFAULT true,
    pausado BOOLEAN DEFAULT false,
    proxima_execucao TIMESTAMP,
    execucoes_realizadas INTEGER DEFAULT 0,
    max_execucoes INTEGER DEFAULT 1,
    
    -- Metadados
    dados_contexto JSONB, -- Dados do contexto da conversa
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Índice único para evitar duplicatas
    UNIQUE(resposta_rapida_id, chat_id, usuario_id)
);

-- Índices para performance
CREATE INDEX IF NOT EXISTS idx_categorias_respostas_usuario ON categorias_respostas(usuario_id);
CREATE INDEX IF NOT EXISTS idx_categorias_respostas_ativo ON categorias_respostas(ativo);

CREATE INDEX IF NOT EXISTS idx_respostas_rapidas_usuario ON respostas_rapidas(usuario_id);
CREATE INDEX IF NOT EXISTS idx_respostas_rapidas_categoria ON respostas_rapidas(categoria_id);
CREATE INDEX IF NOT EXISTS idx_respostas_rapidas_ativo ON respostas_rapidas(ativo, pausado);
CREATE INDEX IF NOT EXISTS idx_respostas_rapidas_agendamento ON respostas_rapidas(agendamento_ativo);

CREATE INDEX IF NOT EXISTS idx_acoes_respostas_resposta ON acoes_respostas(resposta_rapida_id);
CREATE INDEX IF NOT EXISTS idx_acoes_respostas_ordem ON acoes_respostas(resposta_rapida_id, ordem);

CREATE INDEX IF NOT EXISTS idx_execucoes_respostas_usuario ON execucoes_respostas(usuario_id);
CREATE INDEX IF NOT EXISTS idx_execucoes_respostas_chat ON execucoes_respostas(chat_id);
CREATE INDEX IF NOT EXISTS idx_execucoes_respostas_status ON execucoes_respostas(status);
CREATE INDEX IF NOT EXISTS idx_execucoes_respostas_agendado ON execucoes_respostas(agendado_para);

CREATE INDEX IF NOT EXISTS idx_agendamentos_respostas_usuario ON agendamentos_respostas(usuario_id);
CREATE INDEX IF NOT EXISTS idx_agendamentos_respostas_chat ON agendamentos_respostas(chat_id);
CREATE INDEX IF NOT EXISTS idx_agendamentos_respostas_ativo ON agendamentos_respostas(ativo, pausado);
CREATE INDEX IF NOT EXISTS idx_agendamentos_respostas_proxima ON agendamentos_respostas(proxima_execucao);

-- Triggers para updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_categorias_respostas_updated_at BEFORE UPDATE ON categorias_respostas FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_respostas_rapidas_updated_at BEFORE UPDATE ON respostas_rapidas FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_acoes_respostas_updated_at BEFORE UPDATE ON acoes_respostas FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_execucoes_respostas_updated_at BEFORE UPDATE ON execucoes_respostas FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_agendamentos_respostas_updated_at BEFORE UPDATE ON agendamentos_respostas FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
