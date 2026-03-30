FROM golang:1.24-alpine AS builder

ENV GOTOOLCHAIN=auto

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /storemaker cmd/api/main.go

FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata
ENV TZ=America/Sao_Paulo

COPY --from=builder /storemaker /storemaker

EXPOSE 3080

CMD ["/storemaker"]
