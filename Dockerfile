# Utiliza uma imagem oficial do Go como base para a construção
FROM golang:1.22 AS builder

# Define o diretório de trabalho dentro do container
WORKDIR /app

# Copia os arquivos do projeto para o diretório de trabalho
COPY . .

# Compila a aplicação Go para a arquitetura correta
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

# Utiliza uma imagem mais leve para executar a aplicação
FROM alpine:latest

# Define o diretório de trabalho no container final
WORKDIR /kvcli

# Copia o binário compilado da imagem de build
COPY --from=builder /app/main .

# Garante que o binário tenha permissão de execução
RUN chmod +x ./main

# Define o ponto de entrada do container para a execução da aplicação
ENTRYPOINT ["./main"]
