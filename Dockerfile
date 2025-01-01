FROM golang:1.23 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY ./ ./
RUN CGO_ENABLED=0 go build -o ./ged-shovel ./cmd/ged-shovel


FROM scratch
WORKDIR /app
COPY --from=builder /app/ged-shovel ./
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENV USER=appuser
ENTRYPOINT ["./ged-shovel", "run", "--nats=nats://nats:4222"]
