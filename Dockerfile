FROM golang:1.24.1-alpine AS builder
RUN apk add --no-cache make git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/server ./cmd/server/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates
RUN adduser -D -H -h /app appuser
WORKDIR /app

RUN apk add --no-cache gettext

COPY --from=builder /app/server .
COPY .env .

RUN mkdir -p /app/migrations
RUN chown -R appuser:appuser /app
USER appuser

ARG SERVER_PORT=8080
RUN export $(grep -v '^#' .env | xargs) && \
    SERVER_PORT=${SERVER_PORT:-8080} && \
    echo "EXPOSE ${SERVER_PORT}" > /tmp/port_config && \
    echo "CMD [\"./server\"]" >> /tmp/port_config

EXPOSE ${SERVER_PORT}

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD export $(grep -v '^#' .env | xargs) && \
    wget -q --spider http://localhost:${SERVER_PORT}/health || exit 1

CMD export $(grep -v '^#' .env | xargs) && ./server