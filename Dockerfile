# Etap 1: Budowanie (Builder)
FROM golang:1.26-alpine AS builder

RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates && \
    adduser -D -g '' appuser

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o glance .

FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /app/glance /glance

USER appuser

# Dokumentacja portu
EXPOSE 8080/tcp

ENTRYPOINT ["/glance", "--config", "/app/config/glance.yml"]