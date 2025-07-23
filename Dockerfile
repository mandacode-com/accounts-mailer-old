############################
# 1. Build Stage
############################
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

ENV CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o server ./cmd/server

############################
# 2. CA Certificates Stage
############################
FROM alpine AS certs
RUN apk add --no-cache ca-certificates

############################
# 3. Runtime Stage (scratch)
############################
FROM scratch AS prod

# Copy CA certificates
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy built binary
COPY --from=builder /app/server /app/server
COPY --from=builder /app/template /app/template

WORKDIR /app
USER 1001

ENTRYPOINT ["/app/server"]
