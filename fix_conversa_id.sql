-- Script para corrigir o tipo da coluna conversa_id na tabela cards
-- Execute este script diretamente no PostgreSQL

-- 1. Remover constraints de foreign key
ALTER TABLE cards DROP CONSTRAINT IF EXISTS fk_conversas_cards;
ALTER TABLE cards DROP CONSTRAINT IF EXISTS fk_cards_conversa;
ALTER TABLE cards DROP CONSTRAINT IF EXISTS cards_conversa_id_fkey;

-- 2. Limpar dados existentes para evitar conflitos
DELETE FROM cards WHERE conversa_id IS NOT NULL;

-- 3. Alterar o tipo da coluna para VARCHAR
ALTER TABLE cards ALTER COLUMN conversa_id TYPE VARCHAR(255) USING conversa_id::VARCHAR(255);

-- 4. Verificar se a alteração foi bem-sucedida
SELECT column_name, data_type, character_maximum_length 
FROM information_schema.columns 
WHERE table_name = 'cards' AND column_name = 'conversa_id';

-- Resultado esperado: conversa_id | character varying | 255
