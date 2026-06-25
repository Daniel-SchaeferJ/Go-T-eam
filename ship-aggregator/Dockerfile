FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o ship-aggregator .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/ship-aggregator /usr/local/bin/ship-aggregator

ENTRYPOINT ["ship-aggregator"]
