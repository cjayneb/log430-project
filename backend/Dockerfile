# Build stage
FROM golang:alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source code
COPY . .

# Build the binary
RUN go build -o brokerx .

# Run stage
FROM alpine:3.22

WORKDIR /app

# Copy the binary
COPY --from=builder /app/brokerx .

# Environment variables
ENV APP_PORT=8080
ENV DATABASE_URL=root:root@tcp(db:3306)/brokerx

EXPOSE 8080

CMD ["./brokerx"]
