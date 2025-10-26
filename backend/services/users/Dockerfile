# Build stage ==================================
FROM golang:alpine AS builder

# Set module mode on for reproducible builds
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app
RUN apk add --no-cache git

# Copy go.mod/go.sum first to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .
WORKDIR /app
RUN go build -o /app/brokerx-user-service .

# Run stage ====================================
FROM alpine:3.22 AS runtime
WORKDIR /app

COPY --from=builder /app/brokerx-user-service .

EXPOSE 8080
CMD ["./brokerx-user-service"]
