# Etapa 1: Build
FROM golang:1.25-alpine AS builder

# Instalăm git (uneori necesar pentru dependințe go)
RUN apk add --no-cache git

WORKDIR /app

# Copiem fișierele de module PRIMA DATĂ (pentru caching eficient)
COPY go.mod go.sum ./
RUN go mod download

# Copiem restul codului sursă
COPY . .

# Compilăm binarul
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o webhook .

# Etapa 2: Imaginea finală (mică și sigură)
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/webhook .

# Expunem portul HTTPS
EXPOSE 443

ENTRYPOINT ["./webhook"]