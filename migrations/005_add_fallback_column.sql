-- Adicionar coluna fallback na tabela respostas_rapidas
ALTER TABLE respostas_rapidas ADD COLUMN fallback BOOLEAN DEFAULT FALSE;

-- Criar index para consultas por fallback
CREATE INDEX idx_respostas_rapidas_fallback ON respostas_rapidas(fallback);

-- Comentário para explicar o campo
COMMENT ON COLUMN respostas_rapidas.fallback IS 'Indica se esta resposta deve ser usada quando nenhuma outra resposta é encontrada';
