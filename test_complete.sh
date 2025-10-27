#!/bin/bash

echo "======================================"
echo "Rate Limiter - Testes Básicos"
echo "======================================"
echo ""

API_URL="http://localhost:8080"

# Teste 1: Health Check
echo "1. Health Check"
curl -s $API_URL/health | jq
echo ""

# Teste 2: Rate Limiting por IP
echo "2. Rate Limiting por IP (limite: 10 req/s)"
echo "Fazendo 12 requisições..."
for i in {1..12}; do
    status=$(curl -s -o /dev/null -w "%{http_code}" $API_URL)
    echo "  Requisição $i: HTTP $status"
    sleep 0.1
done
echo ""

# Aguardar reset
echo "3. Aguardando reset da janela (2s)..."
sleep 2
echo ""

# Teste 3: Rate Limiting com Token
echo "4. Rate Limiting com Token 'abc123' (limite: 100 req/s)"
echo "Fazendo 5 requisições..."
for i in {1..5}; do
    status=$(curl -s -o /dev/null -w "%{http_code}" -H "API_KEY: abc123" $API_URL)
    echo "  Requisição $i: HTTP $status"
    sleep 0.1
done
echo ""

# Teste 4: Resposta de sucesso
echo "5. Exemplo de resposta de sucesso:"
curl -s -H "API_KEY: test-token" $API_URL | jq
echo ""

# Teste 5: Resposta bloqueada
echo "6. Excedendo limite para gerar bloqueio..."
for i in {1..11}; do
    curl -s -o /dev/null $API_URL
done
echo "Resposta bloqueada:"
curl -s $API_URL | jq
echo ""

echo "======================================"
echo "Testes Concluídos!"
echo "======================================"
