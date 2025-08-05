#!/bin/bash

echo "🧪 Demonstração do Rate Limiter"
echo "================================"
echo ""

echo "1. Testando health check..."
curl -s http://localhost:8080/health
echo ""
echo ""

echo "2. Testando rate limiting por IP (limite: 10 req/s)"
echo "   Fazendo 10 requisições com IP fixo..."
echo ""

for i in {1..10}; do
    echo "   Requisição $i:"
    response=$(curl -s -w "HTTP: %{http_code}" -H "X-Forwarded-For: 192.168.1.297" http://localhost:8080/test)
    echo "   $response"
    echo ""
    sleep 0.1
done

echo "3. 11ª requisição (DEVE SER BLOQUEADA - HTTP 429):"
response=$(curl -s -w "HTTP: %{http_code}" -H "X-Forwarded-For: 192.168.1.297" http://localhost:8080/test)
echo "   $response"
echo ""

echo "4. Testando com IP diferente (deve funcionar):"
response=$(curl -s -w "HTTP: %{http_code}" -H "X-Forwarded-For: 192.168.1.400" http://localhost:8080/test)
echo "   $response"
echo ""

echo "5. Testando rate limiting por token (limite: 100 req/s)"
echo "   Fazendo 5 requisições com token..."
echo ""

for i in {1..5}; do
    echo "   Requisição $i com token:"
    response=$(curl -s -w "HTTP: %{http_code}" -H "API_KEY: test-token-123" http://localhost:8080/test)
    echo "   $response"
    echo ""
    sleep 0.1
done

echo "✅ Demonstração concluída!"
echo ""
echo "💡 Dicas:"
echo "   - Use 'X-Forwarded-For' para simular IPs diferentes"
echo "   - Use 'API_KEY' para testar rate limiting por token"
echo "   - O limite por IP é 10 req/s"
echo "   - O limite por token é 100 req/s"
