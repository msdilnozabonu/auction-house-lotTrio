# ===== stage 1: сборка бинарника =====
FROM golang:1.26-alpine AS build
WORKDIR /src

# сначала только манифесты — кэш слоя зависимостей
COPY go.mod go.sum ./
RUN go mod download

# затем код и статичная сборка
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app ./cmd/server

# ===== stage 2: запуск =====
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
RUN adduser -D -u 10001 appuser

COPY --from=build /app /app
USER appuser
EXPOSE 9999
ENTRYPOINT ["/app"]
