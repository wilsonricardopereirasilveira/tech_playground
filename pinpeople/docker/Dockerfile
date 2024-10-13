# Estágio de compilação
FROM golang:1.18-alpine AS builder

WORKDIR /app

# Copia todo o conteúdo do projeto
COPY . .

# Verifica se go.mod existe, se não, inicializa um novo módulo
RUN if [ ! -f go.mod ]; then go mod init pinpeople; fi

# Baixa as dependências
RUN go mod tidy

# Compila a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

# Estágio final
FROM alpine:latest  

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copia o binário compilado do estágio de compilação
COPY --from=builder /app/main .

# Expõe a porta que a aplicação usa
EXPOSE 8080

# Comando para executar a aplicação
CMD ["./main"]