# ── Builder ──────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

ENV GOTOOLCHAIN=auto

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /storemaker cmd/api/main.go

# ── Runtime (distroless) ─────────────────────────────────
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /storemaker /storemaker

ENV TZ=America/Sao_Paulo

EXPOSE 3080

USER nonroot:nonroot

ENTRYPOINT ["/storemaker"]
