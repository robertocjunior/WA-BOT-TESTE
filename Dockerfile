# Build stage
FROM golang:1.24-bookworm AS builder

# Set working directory
WORKDIR /app

# Install build dependencies (needed for SQLite/CGO)
RUN apt-get update && apt-get install -y gcc libc6-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=1 is required for the SQLite driver
RUN CGO_ENABLED=1 GOOS=linux go build -o bot cmd/bot/main.go

# Final stage
FROM debian:bookworm-slim

WORKDIR /app

# Install runtime dependencies for SQLite and SSL
RUN apt-get update && apt-get install -y ca-certificates libc6 && rm -rf /var/lib/apt/lists/*

# Copy binary from builder
COPY --from=builder /app/bot .

# Create data directory for SQLite
RUN mkdir -p /app/data

# Set environment variables
ENV LOG_LEVEL=info

# Run the bot
CMD ["./bot"]
