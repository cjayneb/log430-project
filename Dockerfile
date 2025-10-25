# Build stage ==================================
FROM golang:alpine AS builder

WORKDIR /app
RUN apk add --no-cache git

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./backend
WORKDIR /app/backend
RUN go build -o /app/brokerx .

# Run stage ====================================
FROM alpine:3.22 AS runtime
WORKDIR /app
COPY frontend ./frontend

COPY backend/adapters/resources ./resources

COPY --from=builder /app/brokerx .
EXPOSE 8080
CMD ["./brokerx"]

# Locust stage =================================
FROM python:3.11-slim AS locust
WORKDIR /app/src
COPY load-tests/ /mnt/locust/

COPY load-tests/requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt

CMD ["locust", "-f", "/mnt/locust/locustfile.py", "--host", "http://nginx:80", "--processes", "-1"]
