# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN apk add --no-cache git && go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o agent-core ./cmd/main.go

# Final stage
FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/agent-core .
COPY --from=builder /app/.env .
COPY --from=builder /app/internal/llm/configs ./internal/llm/configs
COPY --from=builder /app/configs/prompts ./configs/prompts

EXPOSE 8080
CMD ["./agent-core"]
