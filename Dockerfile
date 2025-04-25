# Etapa 1: build
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY main.go ./

RUN go mod tidy

# ✅ Build estático para Cloud Run (amd64)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o hello-app .

# Etapa 2: runtime minimal
FROM alpine:latest

RUN adduser -D appuser
WORKDIR /home/appuser

COPY --from=builder /app/hello-app .

USER appuser

EXPOSE 8080
CMD ["./hello-app"]
