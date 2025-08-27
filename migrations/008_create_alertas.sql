-- Migration: 008_create_alertas
-- Description: Create alertas table for alert management system

CREATE TABLE alertas (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    titulo VARCHAR(255) NOT NULL,
    descricao TEXT,
    tipo VARCHAR(50) NOT NULL CHECK (tipo IN ('sistema', 'usuario', 'seguranca', 'performance', 'integracao')),
    prioridade VARCHAR(50) NOT NULL CHECK (prioridade IN ('baixa', 'media', 'alta', 'critica')),
    status VARCHAR(50) NOT NULL DEFAULT 'ativo' CHECK (status IN ('ativo', 'pausado', 'resolvido')),
    cor VARCHAR(7) NOT NULL DEFAULT '#3B82F6',
    icone VARCHAR(50) NOT NULL DEFAULT 'Bell',
    configuracoes JSONB NOT NULL DEFAULT '{}',
    estatisticas JSONB NOT NULL DEFAULT '{"totalDisparos": 0, "disparosHoje": 0, "ultimoDisparo": null, "taxaResolucao": 0}',
    criado_em TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    atualizado_em TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT alertas_titulo_not_empty CHECK (LENGTH(TRIM(titulo)) > 0)
);

-- Create indexes for better performance
CREATE INDEX idx_alertas_user_id ON alertas(user_id);
CREATE INDEX idx_alertas_status ON alertas(status);
CREATE INDEX idx_alertas_tipo ON alertas(tipo);
CREATE INDEX idx_alertas_prioridade ON alertas(prioridade);
CREATE INDEX idx_alertas_criado_em ON alertas(criado_em);

-- Create alerta_historico table for tracking alert triggers
CREATE TABLE alerta_historico (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alerta_id UUID NOT NULL REFERENCES alertas(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    disparado_em TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    metrica VARCHAR(255),
    valor_detectado TEXT,
    resolvido BOOLEAN DEFAULT FALSE,
    resolvido_em TIMESTAMP WITH TIME ZONE,
    observacoes TEXT
);

-- Create indexes for alerta_historico
CREATE INDEX idx_alerta_historico_alerta_id ON alerta_historico(alerta_id);
CREATE INDEX idx_alerta_historico_user_id ON alerta_historico(user_id);
CREATE INDEX idx_alerta_historico_disparado_em ON alerta_historico(disparado_em);
CREATE INDEX idx_alerta_historico_resolvido ON alerta_historico(resolvido);

-- Create trigger to update atualizado_em automatically
CREATE OR REPLACE FUNCTION update_alertas_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.atualizado_em = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_alertas_updated_at
    BEFORE UPDATE ON alertas
    FOR EACH ROW
    EXECUTE FUNCTION update_alertas_updated_at();

-- Insert some sample data
INSERT INTO alertas (user_id, titulo, descricao, tipo, prioridade, status, cor, icone, configuracoes) VALUES
(
    (SELECT id FROM users LIMIT 1),
    'Tempo de Resposta Alto',
    'Alerta quando o tempo médio de resposta excede 5 minutos',
    'performance',
    'alta',
    'ativo',
    '#EF4444',
    'Clock',
    '{
        "emailNotificacao": true,
        "whatsappNotificacao": false,
        "dashboardNotificacao": true,
        "frequencia": "imediata",
        "destinatarios": ["admin@tappyone.com"],
        "condicoes": [
            {
                "metrica": "tempo_resposta_medio",
                "operador": ">",
                "valor": 300
            }
        ]
    }'
),
(
    (SELECT id FROM users LIMIT 1),
    'Falha na Integração WhatsApp',
    'Monitora falhas de conexão com a API do WhatsApp',
    'integracao',
    'critica',
    'ativo',
    '#DC2626',
    'AlertTriangle',
    '{
        "emailNotificacao": true,
        "whatsappNotificacao": true,
        "dashboardNotificacao": true,
        "frequencia": "imediata",
        "destinatarios": ["admin@tappyone.com"],
        "condicoes": [
            {
                "metrica": "whatsapp_status",
                "operador": "=",
                "valor": "disconnected"
            }
        ]
    }'
);
