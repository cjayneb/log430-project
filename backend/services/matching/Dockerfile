# Build stage ==================================
FROM golang:1.24.9-bookworm AS builder

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

# Copy go.mod/go.sum first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build binary using system librdkafka (ARM64 native)
RUN go build -tags dynamic -o /app/brokerx-matching-service .

# Runtime stage ====================================
FROM debian:bookworm-slim AS runtime

WORKDIR /app

# Install runtime librdkafka
RUN apt-get update && \
    apt-get install -y --no-install-recommends librdkafka1 && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/brokerx-matching-service .

EXPOSE 8080
CMD ["./brokerx-matching-service"]
