.PHONY: help build run test clean docker-build docker-run docker-stop

# Variáveis
APP_NAME=rate-limiter
DOCKER_IMAGE=rate-limiter:latest

# Comando padrão
help: ## Mostra esta ajuda
	@echo "Comandos disponíveis:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Desenvolvimento
build: ## Compila a aplicação
	go build -o bin/$(APP_NAME) ./cmd/server

run: ## Executa a aplicação localmente
	go run ./cmd/server/main.go

test: ## Executa os testes
	go test -v ./...

test-coverage: ## Executa os testes com cobertura
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Remove arquivos de build
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker
docker-build: ## Constrói a imagem Docker
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Executa com docker-compose
	docker-compose up -d

docker-stop: ## Para os containers
	docker-compose down

docker-logs: ## Mostra logs dos containers
	docker-compose logs -f

# Dependências
deps: ## Baixa as dependências
	go mod download
	go mod tidy

# Linting e formatação
fmt: ## Formata o código
	go fmt ./...

lint: ## Executa o linter
	golangci-lint run

# Testes de integração
test-integration: ## Executa testes de integração
	@echo "Iniciando Redis..."
	docker-compose up -d redis
	@echo "Aguardando Redis estar pronto..."
	sleep 5
	@echo "Executando testes..."
	go test -v ./tests/...

# Scripts de teste
test-script: ## Executa o script de teste
	@echo "Certifique-se de que a aplicação está rodando em localhost:8080"
	./scripts/test_rate_limiter.sh

# Desenvolvimento completo
dev: ## Inicia ambiente de desenvolvimento completo
	@echo "Iniciando ambiente de desenvolvimento..."
	docker-compose up -d redis
	@echo "Aguardando Redis estar pronto..."
	sleep 5
	@echo "Iniciando aplicação..."
	go run ./cmd/server/main.go

# Monitoramento
monitor: ## Monitora a aplicação
	@echo "Status dos containers:"
	docker-compose ps
	@echo ""
	@echo "Logs da aplicação:"
	docker-compose logs app --tail=20
	@echo ""
	@echo "Logs do Redis:"
	docker-compose logs redis --tail=10

# Utilitários
check-env: ## Verifica se o arquivo .env existe
	@if [ ! -f .env ]; then \
		echo "Arquivo .env não encontrado. Copiando de .env.example..."; \
		cp .env.example .env; \
		echo "Arquivo .env criado. Edite conforme necessário."; \
	else \
		echo "Arquivo .env encontrado."; \
	fi

setup: ## Configura o ambiente de desenvolvimento
	@echo "Configurando ambiente de desenvolvimento..."
	$(MAKE) check-env
	$(MAKE) deps
	@echo "Ambiente configurado!"

# Instalação de ferramentas
install-tools: ## Instala ferramentas de desenvolvimento
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/go-delve/delve/cmd/dlv@latest

# Debug
debug: ## Executa a aplicação em modo debug
	dlv debug ./cmd/server/main.go

# Rate Limiter - Limpeza
redis-clear-all: ## Limpa todos os rate limits do Redis
	@echo "Limpando todos os rate limits do Redis..."
	docker-compose exec redis redis-cli FLUSHDB
	@echo "✅ Rate limits limpos!"

redis-clear-ip: ## Limpa rate limits de um IP específico
	@echo "Uso: make redis-clear-ip IP=192.168.1.100"
	@if [ -z "$(IP)" ]; then \
		echo "❌ Erro: Especifique um IP. Exemplo: make redis-clear-ip IP=192.168.1.100"; \
		exit 1; \
	fi
	@echo "Limpando rate limits do IP $(IP)..."
	docker-compose exec redis redis-cli DEL "rate:ip:$(IP)"
	docker-compose exec redis redis-cli DEL "block:ip:$(IP)"
	@echo "✅ Rate limits do IP $(IP) limpos!"

redis-clear-token: ## Limpa rate limits de um token específico
	@echo "Uso: make redis-clear-token TOKEN=test-token-123"
	@if [ -z "$(TOKEN)" ]; then \
		echo "❌ Erro: Especifique um TOKEN. Exemplo: make redis-clear-token TOKEN=test-token-123"; \
		exit 1; \
	fi
	@echo "Limpando rate limits do token $(TOKEN)..."
	docker-compose exec redis redis-cli DEL "rate:token:$(TOKEN)"
	docker-compose exec redis redis-cli DEL "block:token:$(TOKEN)"
	@echo "✅ Rate limits do token $(TOKEN) limpos!"

redis-list: ## Lista todas as chaves de rate limit no Redis
	@echo "Listando chaves de rate limit no Redis..."
	docker-compose exec redis redis-cli KEYS "*rate*"
	docker-compose exec redis redis-cli KEYS "*block*"

redis-info: ## Mostra informações sobre rate limits
	@echo "Informações sobre rate limits no Redis:"
	@echo "Chaves de rate limit:"
	docker-compose exec redis redis-cli KEYS "*rate*" | wc -l
	@echo "Chaves de bloqueio:"
	docker-compose exec redis redis-cli KEYS "*block*" | wc -l
