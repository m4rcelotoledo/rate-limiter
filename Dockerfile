FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copia os arquivos de dependências
COPY go.mod go.sum ./

# Baixa as dependências
RUN go mod download

# Copia o código fonte
COPY . .

# Compila a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Imagem final
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copia o binário e arquivos de configuração
COPY --from=builder /app/main .
COPY .env.example .env
COPY static ./static

# Expõe a porta
EXPOSE 8080

# Comando para executar a aplicação
CMD ["./main"]
