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
COPY --from=builder /app/server .
RUN mkdir -p /app/migrations
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1
CMD ["./server"]
