# Build stage
FROM golang:1.25.7-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build Go application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main .

# Run stage
FROM alpine:latest

# Install basic runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy built binary from builder
COPY --from=builder /app/main .

# Expose Go API port
EXPOSE 8080

# Run the app
CMD ["./main"]
