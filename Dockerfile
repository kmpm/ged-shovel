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
COPY --from=builder /etc/services /etc/services
ENV USER=appuser
ENTRYPOINT ["./ged-shovel"]
CMD ["run", "--metrics=:2112", "--nats=nats://nats:4222"]
