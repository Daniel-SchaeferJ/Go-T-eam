FROM golang:1.25-alpine AS builder

# This service requires the following environment variables at runtime:
# - AISSTREAM_API_KEY
# - HF_TOKEN
# - HF_REPO

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ship-aggregator .

FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/ship-aggregator /usr/local/bin/ship-aggregator

# Ensure binary is executable
RUN chmod +x /usr/local/bin/ship-aggregator

ENTRYPOINT ["/usr/local/bin/ship-aggregator"]
