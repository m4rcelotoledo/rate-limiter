#!/bin/bash

# Script para testar o rate limiter
# Certifique-se de que a aplicação está rodando em localhost:8080

echo "🧪 Testando Rate Limiter"
echo "=========================="

# Função para fazer requisição e mostrar resultado
make_request() {
    local url=$1
    local headers=$2
    local description=$3

    echo -n "$description: "

    if [ -z "$headers" ]; then
        response=$(curl -s -w "%{http_code}" -H "X-Forwarded-For: 192.168.1.100" -o /tmp/response.json http://localhost:8080$url)
    else
        response=$(curl -s -w "%{http_code}" -H "$headers" -H "X-Forwarded-For: 192.168.1.100" -o /tmp/response.json http://localhost:8080$url)
    fi

    http_code=${response: -3}

    if [ "$http_code" = "200" ]; then
        echo "✅ Sucesso (HTTP $http_code)"
        if [ -f /tmp/response.json ]; then
            echo "   Resposta: $(cat /tmp/response.json)"
        fi
    elif [ "$http_code" = "429" ]; then
        echo "🚫 Rate Limit Excedido (HTTP $http_code)"
        if [ -f /tmp/response.json ]; then
            echo "   Erro: $(cat /tmp/response.json)"
        fi
    else
        echo "❌ Erro (HTTP $http_code)"
        if [ -f /tmp/response.json ]; then
            echo "   Erro: $(cat /tmp/response.json)"
        fi
    fi

    echo ""
}

# Função para fazer requisição com IP específico
make_request_with_ip() {
    local url=$1
    local headers=$2
    local ip=$3
    local description=$4

    echo -n "$description: "

    if [ -z "$headers" ]; then
        response=$(curl -s -w "%{http_code}" -H "X-Forwarded-For: $ip" -o /tmp/response.json http://localhost:8080$url)
    else
        response=$(curl -s -w "%{http_code}" -H "$headers" -H "X-Forwarded-For: $ip" -o /tmp/response.json http://localhost:8080$url)
    fi

    http_code=${response: -3}

    if [ "$http_code" = "200" ]; then
        echo "✅ Sucesso (HTTP $http_code)"
        if [ -f /tmp/response.json ]; then
            echo "   Resposta: $(cat /tmp/response.json)"
        fi
    elif [ "$http_code" = "429" ]; then
        echo "🚫 Rate Limit Excedido (HTTP $http_code)"
        if [ -f /tmp/response.json ]; then
            echo "   Erro: $(cat /tmp/response.json)"
        fi
    else
        echo "❌ Erro (HTTP $http_code)"
        if [ -f /tmp/response.json ]; then
            echo "   Erro: $(cat /tmp/response.json)"
        fi
    fi

    echo ""
}

# Teste 1: Health check
echo "1. Testando Health Check"
make_request "/health" "" "Health Check"

# Teste 2: Requisições sem token (limitação por IP)
echo "2. Testando Limitação por IP"
echo "   Fazendo 10 requisições consecutivas com IP fixo..."

for i in {1..10}; do
    make_request_with_ip "/test" "" "192.168.1.200" "Requisição $i (IP: 192.168.1.200)"
    sleep 0.1
done

echo "   Fazendo 11ª requisição (deve ser bloqueada)..."
make_request_with_ip "/test" "" "192.168.1.200" "Requisição 11 (IP: 192.168.1.200 - deve ser bloqueada)"

# Teste 3: IP diferente (deve funcionar)
echo "3. Testando IP Diferente"
make_request_with_ip "/test" "" "192.168.1.300" "IP Diferente (deve funcionar)"

# Teste 4: Requisições com token (limitação por token)
echo "4. Testando Limitação por Token"
echo "   Fazendo 10 requisições consecutivas com token..."

for i in {1..100}; do
    make_request_with_ip "/test" "API_KEY: test-token-123" "192.168.1.400" "Requisição $i (com token)"
    sleep 0.1
done

echo "   Fazendo 101ª requisição (deve ser bloqueada)..."
make_request_with_ip "/test" "API_KEY: test-token-123" "192.168.1.400" "Requisição 101 (com token - deve ser bloqueada)"

# Teste 5: Diferentes tokens
echo "5. Testando Diferentes Tokens"
make_request_with_ip "/test" "API_KEY: token-abc" "192.168.1.500" "Token ABC"
make_request_with_ip "/test" "API_KEY: token-def" "192.168.1.600" "Token DEF"
make_request_with_ip "/test" "API_KEY: token-ghi" "192.168.1.700" "Token GHI"

# Teste 6: Headers de rate limit
echo "6. Verificando Headers de Rate Limit"
echo "   Fazendo requisição e verificando headers..."

echo "   Requisição 1 (deve ter headers de rate limit):"
response=$(curl -s -D - -H "X-Forwarded-For: 192.168.1.800" http://localhost:8080/test | head -10)
echo "   Headers de Rate Limit:"
echo "$response" | grep -E "X-Ratelimit-" | sed 's/^/   /'

echo ""
echo "   Requisição 2 (deve mostrar remaining diminuído):"
response=$(curl -s -D - -H "X-Forwarded-For: 192.168.1.800" http://localhost:8080/test | head -10)
echo "   Headers de Rate Limit:"
echo "$response" | grep -E "X-Ratelimit-" | sed 's/^/   /'

echo ""
echo "   Requisição 3 (deve mostrar remaining ainda menor):"
response=$(curl -s -D - -H "X-Forwarded-For: 192.168.1.800" http://localhost:8080/test | head -10)
echo "   Headers de Rate Limit:"
echo "$response" | grep -E "X-Ratelimit-" | sed 's/^/   /'

echo ""
echo "✅ Testes concluídos!"
echo ""
echo "💡 Dicas:"
echo "   - Para testar com diferentes IPs, use: curl -H 'X-Forwarded-For: 192.168.1.1' http://localhost:8080/test"
echo "   - Para testar com diferentes tokens, use: curl -H 'API_KEY: seu-token' http://localhost:8080/test"
echo "   - Para ver logs detalhados, use: docker-compose logs -f app"
