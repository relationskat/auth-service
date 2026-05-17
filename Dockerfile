FROM golang:1.25-alpine AS builder

WORKDIR /build

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
  -trimpath -ldflags="-s -w" \
  -o /build/auth-service ./cmd/app

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 appuser

WORKDIR /app

COPY --from=builder /build/auth-service /app/auth-service
COPY config /app/config
COPY infra/keys /app/keys

USER appuser

EXPOSE 50051

ENTRYPOINT ["/app/auth-service"]
