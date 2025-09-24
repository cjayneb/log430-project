# Build stage ==================================
FROM golang:alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Copy Go modules and download dependencies
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/ ./backend

# Build the binary
WORKDIR /app/backend
RUN go build -o /app/brokerx .

# Run stage ====================================
FROM alpine:3.22

WORKDIR /app

# Copy frontend for static serving
COPY frontend ./frontend

# Copy the binary
COPY --from=builder /app/brokerx .

# Environment variables
ENV APP_PORT=8080
ENV DATABASE_URL=root:root@tcp(db:3306)/brokerx

EXPOSE 8080

CMD ["./brokerx"]
