# Etapa de build
FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o clima-por-cep .

# Imagem final mínima
FROM scratch

# Certificados para as chamadas HTTPS ao ViaCEP e WeatherAPI
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/clima-por-cep /clima-por-cep

EXPOSE 8080

ENTRYPOINT ["/clima-por-cep"]
