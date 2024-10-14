# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copiar go.mod e go.sum (se existir)
COPY go.mod ./
COPY go.sum* ./

# Baixar dependências
RUN go mod download

# Copiar o restante do código fonte
COPY . .

# Compilar a aplicação principal
RUN go build -o main ./cmd/api

# Compilar o importador
RUN go build -o import ./cmd/import

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copiar os executáveis compilados
COPY --from=builder /app/main .
COPY --from=builder /app/import .

# Copiar o arquivo CSV
COPY data/employees.csv ./data/

EXPOSE 8080

# Comando padrão para executar a aplicação principal
CMD ["./main"]