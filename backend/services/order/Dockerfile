# Build stage ==================================
FROM golang:1.24.3-bookworm AS builder

ENV CGO_ENABLED=1
ENV PKG_CONFIG_PATH=/usr/lib/pkgconfig

WORKDIR /app
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        build-essential \
        git \
        pkg-config \
        librdkafka-dev && \
    rm -rf /var/lib/apt/lists/*

# Copy go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary using system librdkafka (ARM64 native)
RUN go build -tags dynamic -o /app/brokerx-order-service .

# Runtime stage ====================================
FROM debian:bookworm-slim AS runtime

WORKDIR /app

# Install runtime librdkafka
RUN apt-get update && \
    apt-get install -y --no-install-recommends librdkafka1 && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/brokerx-order-service .

EXPOSE 8080
CMD ["./brokerx-order-service"]
