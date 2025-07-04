# Build stage
FROM golang:1.23.10-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and ca-certificates (needed for fetching dependencies)
RUN apk add --no-cache git=2.47.1-r0 ca-certificates=20241010-r0 tzdata=2024b-r1

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.version=${VERSION}" \
    -a -installsuffix cgo \
    -o golanggraph \
    ./cmd/golanggraph

# Final stage
FROM alpine:3.20.3

# Install ca-certificates for HTTPS requests and create non-root user
RUN apk --no-cache add ca-certificates=20241010-r0 tzdata=2024b-r1 && \
    addgroup -g 1001 -S golanggraph && \
    adduser -u 1001 -S golanggraph -G golanggraph

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/golanggraph .

# Create directories for optional files and change ownership to non-root user
RUN mkdir -p ./configs ./docs && \
    chown -R golanggraph:golanggraph /app

# Switch to non-root user
USER golanggraph

# Expose port (adjust as needed)
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./golanggraph health || exit 1

# Run the binary
ENTRYPOINT ["./golanggraph"]
CMD ["serve"]
