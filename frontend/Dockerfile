FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o frontend-service

FROM alpine:3.22
WORKDIR /app
COPY --from=builder /app/frontend-service .
COPY public ./public
EXPOSE 8081
CMD ["./frontend-service"]
