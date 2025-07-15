# GoLangGraph Go-Agents-Simple Deployment
# Build from the project root to include GoLangGraph dependencies

FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git=2.45.2-r0

WORKDIR /build

# Copy the entire GoLangGraph project (needed for replace directive)
COPY ../../../ .

# Set working directory to our specific example
WORKDIR /build/examples/10-ideation-agents/go-agents-simple

# Build with GOWORK disabled to use module mode
RUN GOWORK=off go mod tidy && GOWORK=off go mod download
RUN CGO_ENABLED=0 GOOS=linux GOWORK=off go build -o main .

# Final stage - minimal runtime image
FROM alpine:3.20.2

# Install runtime dependencies
RUN apk --no-cache add ca-certificates=20240705-r0 wget=1.24.5-r0 curl=8.9.0-r0

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /build/examples/10-ideation-agents/go-agents-simple/main .

# Create non-root user for security
RUN adduser -D -s /bin/sh appuser
USER appuser

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./main"]
