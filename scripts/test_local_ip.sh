#!/bin/bash

echo "🧪 Testando Rate Limiter com IP Local"
echo "======================================"
echo ""

echo "1. Verificando IP local..."
LOCAL_IP=$(ip route get 1.1.1.1 | awk '{print $7}' | head -1)
echo "   IP Local: $LOCAL_IP"
echo ""

echo "2. Testando rate limiting com IP local..."
echo "   Fazendo 10 requisições..."

for i in {1..10}; do
    echo "   Requisição $i:"
    response=$(curl -s -w "HTTP: %{http_code}" -H "X-Forwarded-For: $LOCAL_IP" -o /tmp/response.json http://localhost:8080/test)
    http_code=${response: -3}

    if [ "$http_code" = "200" ]; then
        echo "   ✅ Sucesso (HTTP $http_code)"
        if [ -f /tmp/response.json ]; then
            echo "   Resposta: $(cat /tmp/response.json)"
        fi
    elif [ "$http_code" = "429" ]; then
        echo "   🚫 Rate Limit Excedido (HTTP $http_code)"
        if [ -f /tmp/response.json ]; then
            echo "   Erro: $(cat /tmp/response.json)"
        fi
    else
        echo "   ❌ Erro (HTTP $http_code)"
        if [ -f /tmp/response.json ]; then
            echo "   Erro: $(cat /tmp/response.json)"
        fi
    fi
    echo ""
    sleep 0.1
done

echo "3. 11ª requisição (DEVE SER BLOQUEADA):"
response=$(curl -s -w "HTTP: %{http_code}" -H "X-Forwarded-For: $LOCAL_IP" -o /tmp/response.json http://localhost:8080/test)
http_code=${response: -3}

if [ "$http_code" = "200" ]; then
    echo "   ✅ Sucesso (HTTP $http_code)"
    if [ -f /tmp/response.json ]; then
        echo "   Resposta: $(cat /tmp/response.json)"
    fi
elif [ "$http_code" = "429" ]; then
    echo "   🚫 Rate Limit Excedido (HTTP $http_code)"
    if [ -f /tmp/response.json ]; then
        echo "   Erro: $(cat /tmp/response.json)"
    fi
else
    echo "   ❌ Erro (HTTP $http_code)"
    if [ -f /tmp/response.json ]; then
        echo "   Erro: $(cat /tmp/response.json)"
    fi
fi
echo ""

echo "4. Testando com IP diferente (deve funcionar):"
response=$(curl -s -w "HTTP: %{http_code}" -H "X-Forwarded-For: 192.168.1.999" -o /tmp/response.json http://localhost:8080/test)
http_code=${response: -3}

if [ "$http_code" = "200" ]; then
    echo "   ✅ Sucesso (HTTP $http_code)"
    if [ -f /tmp/response.json ]; then
        echo "   Resposta: $(cat /tmp/response.json)"
    fi
elif [ "$http_code" = "429" ]; then
    echo "   🚫 Rate Limit Excedido (HTTP $http_code)"
    if [ -f /tmp/response.json ]; then
        echo "   Erro: $(cat /tmp/response.json)"
    fi
else
    echo "   ❌ Erro (HTTP $http_code)"
    if [ -f /tmp/response.json ]; then
        echo "   Erro: $(cat /tmp/response.json)"
    fi
fi
echo ""

echo "✅ Teste concluído!"
echo ""
echo "💡 Dicas:"
echo "   - Este teste usa o IP real da sua máquina"
echo "   - Para testar sem headers, use o IP real da sua máquina"
echo "   - O rate limiter funciona por IP, então cada IP tem seu próprio limite"
