# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Set GOPROXY for faster and more reliable module downloads
ENV GOPROXY=https://proxy.golang.org,direct

# Download dependencies first for better caching
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Copy the source code
COPY . .

# Build the API and Seeder binaries using BuildKit cache for faster rebuilds
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS=linux go build -o /app/bin/api ./cmd/api/main.go && \
    CGO_ENABLED=0 GOOS=linux go build -o /app/bin/seeder ./cmd/seeder/main.go

# Run stage
FROM alpine:latest

# Install ca-certificates for secure connections if needed
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binaries from the builder stage
COPY --from=builder /app/bin/api .
COPY --from=builder /app/bin/seeder .

# Expose the API port
EXPOSE 8080

# Default command starts the API
CMD ["./api"]
