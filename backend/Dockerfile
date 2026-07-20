# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /invoice-api .

# Run stage
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /invoice-api .

EXPOSE 8080

CMD ["./invoice-api"]
