# Etapa 1: Build
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copiem fișierele de dependențe
COPY go.mod go.sum ./
RUN go mod download

# Copiem codul sursă
COPY . .

# Compilăm binarul static
RUN CGO_ENABLED=0 GOOS=linux go build -o dispatcher .

# Etapa 2: Imaginea finală de rulare
FROM alpine:latest

# Instalăm certificatele pentru HTTPS și AWS CLI (necesar pentru autentificare EKS)
RUN apk add --no-cache ca-certificates aws-cli

WORKDIR /root/

# Copiem binarul din etapa de build
COPY --from=builder /app/dispatcher .

# Comanda de start
CMD ["./dispatcher"]