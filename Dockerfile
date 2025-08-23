# Multi-stage build for POSIX System MCP Server

# Build stage
FROM golang:1.21-alpine AS builder

# Install git (needed for go modules)
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o posix-system-mcp .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh mcpuser

# Set working directory
WORKDIR /home/mcpuser

# Copy binary from builder
COPY --from=builder /app/posix-system-mcp .

# Change ownership to non-root user
RUN chown mcpuser:mcpuser posix-system-mcp

# Switch to non-root user
USER mcpuser

# Expose port (if needed for future web interface)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./posix-system-mcp --version || exit 1

# Run the binary
CMD ["./posix-system-mcp"]
