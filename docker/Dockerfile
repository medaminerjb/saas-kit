FROM golang:1.26.1-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/bin/saaskit ./cmd/saaskit

# ─── Runtime ───────────────────────────────────────────
FROM gcr.io/distroless/static-debian12

COPY --from=builder /build/bin/saaskit /saaskit
COPY --from=builder /build/migrations /migrations

EXPOSE 8080

ENTRYPOINT ["/saaskit"]
