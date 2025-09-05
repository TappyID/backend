#!/bin/bash

echo "🚀 Iniciando debug para inserção de usuários..."
echo "================================="

SERVER="root@159.65.34.199"

# Verifica se os containers estão rodando
echo "📋 Verificando containers..."
ssh $SERVER "docker ps | grep backend"

echo ""
echo "🔗 Testando conexão com banco..."
ssh $SERVER "docker exec backend-postgres-1 psql -U postgres -d tappyone -c 'SELECT version();'" 2>&1

echo ""
echo "📊 Verificando tabelas existentes..."
ssh $SERVER "docker exec backend-postgres-1 psql -U postgres -d tappyone -c '\dt'" 2>&1

echo ""
echo "👥 Verificando usuários existentes..."
ssh $SERVER "docker exec backend-postgres-1 psql -U postgres -d tappyone -c 'SELECT id, email, nome FROM usuarios LIMIT 5;'" 2>&1

echo ""
echo "🔍 Verificando arquivos no container backend..."
ssh $SERVER "docker exec backend-backend-1 ls -la /app/" 2>&1

echo ""
echo "🔧 Executando create-users..."
ssh $SERVER "docker exec backend-backend-1 /bin/sh -c \"cd /app && if [ -f ./create-users ]; then ./create-users; else echo 'create-users não encontrado'; fi\"" 2>&1

echo ""
echo "✅ Verificando usuários após inserção..."
ssh $SERVER "docker exec backend-postgres-1 psql -U postgres -d tappyone -c 'SELECT id, email, nome FROM usuarios LIMIT 10;'" 2>&1

echo ""
echo "🧪 Testando login API..."
curl -X POST http://159.65.34.199:8081/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@tappy.id", "senha": "admin123"}' \
  -w "\nStatus: %{http_code}\n"

echo ""
echo "================================="
echo "✨ Debug concluído!"
