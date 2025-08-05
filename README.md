# Rate Limiter em Go

Este projeto implementa um rate limiter em Go que pode ser configurado para limitar o número máximo de requisições por segundo com base em um endereço IP específico ou em um token de acesso.

## Funcionalidades

- **Limitação por IP**: Restringe o número de requisições recebidas de um único endereço IP
- **Limitação por Token**: Limita as requisições baseadas em um token de acesso único
- **Prioridade do Token**: As configurações de limite do token se sobrepõem às do IP
- **Middleware**: Pode ser injetado como middleware em um servidor web
- **Configuração Flexível**: Configurações via variáveis de ambiente ou arquivo .env
- **Persistência Redis**: Armazena informações de limite no Redis
- **Strategy Pattern**: Permite trocar facilmente o Redis por outro mecanismo de persistência
- **Interface Web**: Demonstração interativa via browser
- **Scripts de Teste**: Testes automatizados e demonstrações

## Arquitetura

```
rate-limiter/
├── cmd/server/         # Servidor principal
├── internal/
│   ├── config/         # Configurações
│   ├── limiter/        # Lógica do rate limiter
│   ├── middleware/     # Middleware para Gin
│   └── storage/        # Strategy para persistência
├── scripts/            # Scripts de teste e demonstração
├── static/             # Interface web
├── tests/              # Testes de integração
├── Dockerfile          # Containerização
├── docker-compose.yml  # Orquestração
└── README.md           # Documentação
```

## Configuração

### Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto (copiando de `.env.example`) ou configure as seguintes variáveis:

```env
# Configurações do Rate Limiter
RATE_LIMIT_IP_REQUESTS_PER_SECOND=10
RATE_LIMIT_IP_BLOCK_DURATION_SECONDS=300
RATE_LIMIT_TOKEN_REQUESTS_PER_SECOND=100
RATE_LIMIT_TOKEN_BLOCK_DURATION_SECONDS=600

# Configurações do Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Configurações do Servidor
SERVER_PORT=8080
```

### Explicação das Configurações

- `RATE_LIMIT_IP_REQUESTS_PER_SECOND`: Número máximo de requisições por segundo por IP
- `RATE_LIMIT_IP_BLOCK_DURATION_SECONDS`: Tempo de bloqueio em segundos quando o limite por IP é excedido
- `RATE_LIMIT_TOKEN_REQUESTS_PER_SECOND`: Número máximo de requisições por segundo por token
- `RATE_LIMIT_TOKEN_BLOCK_DURATION_SECONDS`: Tempo de bloqueio em segundos quando o limite por token é excedido

## Como Usar

### 1. Executando com Docker Compose

```bash
# Clone o repositório
git clone git@github.com:m4rcelotoledo/rate-limiter.git
cd rate-limiter

# Execute com docker-compose
docker-compose up -d
```

### 2. Executando Localmente

```bash
# Instale as dependências
go mod download

# Execute o Redis (se não estiver rodando)
docker run -d -p 6379:6379 redis:8-alpine

# Execute a aplicação
go run cmd/server/main.go
```

### 3. Testando a API

#### Requisição sem Token (limitação por IP)
```bash
curl http://localhost:8080/test
```

#### Requisição com Token (limitação por token)
```bash
curl -H "API_KEY: abc123" http://localhost:8080/test
```

## Endpoints

- `GET /`: Informações da API
- `GET /health`: Health check
- `GET /static/`: Interface web de demonstração
- `GET /test`: Endpoint de teste
- `POST /test`: Endpoint de teste (POST)

## Headers de Resposta

O rate limiter adiciona os seguintes headers nas respostas:

- `X-RateLimit-Limit`: Limite máximo de requisições
- `X-RateLimit-Remaining`: Requisições restantes
- `X-RateLimit-Reset`: Timestamp de reset do limite

## Resposta de Erro

Quando o limite é excedido, a API retorna:

```json
{
  "error": "you have reached the maximum number of requests or actions allowed within a certain time frame"
}
```

Com status HTTP 429 (Too Many Requests).

## Scripts de Teste

### Teste Completo
```bash
# Teste abrangente com IPs fixos
./scripts/test_rate_limiter.sh
```

### Teste com IP Local
```bash
# Teste usando o IP real da sua máquina
./scripts/test_local_ip.sh
```

### Demonstração Rápida
```bash
# Demonstração simples do rate limiter
./scripts/demo_rate_limiter.sh
```

### Interface Web
Acesse `http://localhost:8080/static/` para uma interface web interativa que permite:
- Testar rate limiting por IP
- Testar rate limiting por token
- Ver headers de rate limit
- Contadores visuais
- Teste rápido automático

## Comandos de Limpeza

### Limpar Rate Limits do Redis

```bash
# Limpar todos os rate limits
make redis-clear-all

# Limpar rate limits de IP específico
make redis-clear-ip IP=192.168.1.100

# Limpar rate limits de token específico
make redis-clear-token TOKEN=test-token-123

# Listar chaves de rate limit
make redis-list

# Ver informações sobre rate limits
make redis-info
```

## Estratégia de Storage

O projeto implementa o padrão Strategy para permitir diferentes implementações de storage:

### Interface StorageStrategy

```go
type StorageStrategy interface {
    Increment(ctx context.Context, key string, expiration time.Duration) (int64, error)
    Get(ctx context.Context, key string) (int64, error)
    Set(ctx context.Context, key string, value int64, expiration time.Duration) error
    Exists(ctx context.Context, key string) (bool, error)
    Delete(ctx context.Context, key string) error
    Close() error
}
```

### Implementação Redis

A implementação atual usa Redis, mas você pode facilmente criar outras implementações (ex: in-memory, PostgreSQL, etc.).

## Testes

### Executando todos os Testes Unitários

```bash
go test -v ./...
```

### Executando apenas os Testes do Rate-Limiter

```bash
go test ./internal/limiter/...
```

### Executando Testes de Integração

```bash
# Certifique-se de que o Redis está rodando
docker-compose up -d redis

# Execute os testes
# Esses testes falharão caso sejam executados em sequência, uma vez que os dados estarão no Redis, sendo liberandos novamente apenas em 60 e 120 segundos depois.
go test ./tests/integration_test.go
```

### Executando Scripts de Teste

```bash
# Teste completo
./scripts/test_rate_limiter.sh

# Teste com IP local
./scripts/test_local_ip.sh

# Demonstração rápida
./demo_rate_limiter.sh
```

## Exemplos de Uso

### 1. Limitação por IP

Suponha que o rate limiter esteja configurado para permitir no máximo 10 requisições por segundo por IP:

```bash
# Primeiras 10 requisições (sucesso)
for i in {1..10}; do
  curl -H "X-Forwarded-For: 192.168.1.100" http://localhost:8080/test
done

# 11ª requisição (bloqueada)
curl -H "X-Forwarded-For: 192.168.1.100" http://localhost:8080/test
# Retorna: 429 Too Many Requests
```

### 2. Limitação por Token

Se um token `abc123` tiver um limite configurado de 100 requisições por segundo:

```bash
# Primeiras 100 requisições (sucesso)
for i in {1..100}; do
  curl -H "API_KEY: abc123" http://localhost:8080/test
done

# 101ª requisição (bloqueada)
curl -H "API_KEY: abc123" http://localhost:8080/test
# Retorna: 429 Too Many Requests
```

### 3. Prioridade do Token sobre IP

Se o limite por IP é de 10 req/s e o de um token é de 100 req/s:

```bash
# Requisição sem token: usa limite de IP (10 req/s)
curl http://localhost:8080/test

# Requisição com token: usa limite do token (100 req/s)
curl -H "API_KEY: abc123" http://localhost:8080/test
```

### 4. Teste via Browser

Acesse `http://localhost:8080/static/` para uma interface web que permite:
- Testar rate limiting interativamente
- Ver contadores em tempo real
- Analisar headers de rate limit
- Fazer testes rápidos

## Desenvolvimento

### Estrutura do Código

- **Config**: Carrega configurações de variáveis de ambiente
- **Limiter**: Lógica principal do rate limiter (separada do middleware)
- **Middleware**: Integração com o framework web (Gin)
- **Storage**: Strategy pattern para persistência

### Adicionando Novas Implementações de Storage

1. Implemente a interface `StorageStrategy`
2. Crie uma função construtora (ex: `NewPostgreSQLStorage`)
3. Use a nova implementação no `main.go`

### Adicionando Novos Frameworks Web

1. Crie um novo middleware que implemente a lógica do rate limiter
2. Use a mesma instância do `RateLimiter` para manter a lógica separada

## Monitoramento

O rate limiter adiciona headers informativos em todas as respostas:

- `X-RateLimit-Limit`: Limite máximo
- `X-RateLimit-Remaining`: Requisições restantes
- `X-RateLimit-Reset`: Timestamp de reset

## Troubleshooting

### Redis não conecta

```bash
# Verifique se o Redis está rodando
docker ps | grep redis

# Teste a conexão
redis-cli ping
```

### Aplicação não inicia

```bash
# Verifique as variáveis de ambiente
cat .env

# Verifique os logs
docker-compose logs app
```

### Rate limiter não funciona

```bash
# Verifique as configurações
curl -v http://localhost:8080/health

# Teste com diferentes IPs/tokens
curl -H "X-Forwarded-For: 192.168.1.1" http://localhost:8080/test
curl -H "API_KEY: test123" http://localhost:8080/test
```

### Testes não funcionam

```bash
# Use IPs fixos para testes
curl -H "X-Forwarded-For: 192.168.1.100" http://localhost:8080/test

# Execute os scripts de teste
./scripts/test_rate_limiter.sh
```

## Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/AmazingFeature`)
3. Commit suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. Push para a branch (`git push origin feature/AmazingFeature`)
5. Abra um Pull Request

## Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.
