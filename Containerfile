# =============================================================================
# Multi-Stage Production Containerfile for Go Application
# =============================================================================
# Compatible with: Podman 4.0+, Docker 20.10+
# Target image size: <50MB
# Security: Non-root user execution (UID 1000), distroless base
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Builder - Build Go binary
# -----------------------------------------------------------------------------
FROM golang:1.23-alpine AS builder

# Install build dependencies
# git: Required for fetching Go modules from VCS
# ca-certificates: SSL/TLS certificate validation for module downloads
RUN apk add --no-cache \
    git=2.45.2-r0 \
    ca-certificates=20240705-r0

# Set working directory for build
WORKDIR /build

# Copy dependency files first for better layer caching
# Go modules are cached unless go.mod/go.sum change
COPY go.mod go.sum ./

# Download dependencies
# Separate layer for dependency caching - only re-runs if go.mod/go.sum change
RUN go mod download && go mod verify

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/

# Build binary with optimizations
# CGO_ENABLED=0: Static binary, no C dependencies
# -ldflags="-s -w": Strip debug info and symbol table (reduces size)
# -trimpath: Remove file system paths from binary (security)
# -tags=netgo: Use pure Go network stack
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
    -trimpath \
    -tags=netgo \
    -o /build/bin/api \
    ./cmd/api

# Verify binary is statically linked
# hadolint ignore=DL4006
RUN ldd /build/bin/api 2>&1 | grep -q "not a dynamic executable"

# -----------------------------------------------------------------------------
# Stage 2: Production - Minimal runtime image
# -----------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS production

# Distroless image benefits:
# - No shell, package manager, or unnecessary binaries (smaller attack surface)
# - Minimal CVE exposure
# - Already configured with non-root user (UID 65532)
# - CA certificates included for HTTPS

# Set working directory
WORKDIR /app

# Copy binary from builder stage
# Distroless uses UID 65532 (nonroot user) by default
COPY --from=builder --chown=65532:65532 /build/bin/api /app/api

# Copy optional configuration files if needed
# COPY --from=builder --chown=65532:65532 /build/configs /app/configs

# Expose application port
# Port 8080: Standard HTTP port for containerized apps
EXPOSE 8080

# Health check configuration
# Interval: Check every 30 seconds
# Timeout: Fail if health check takes >10 seconds
# Start period: Wait 40 seconds after container start before first check
# Retries: Mark unhealthy after 3 consecutive failures
# Note: Distroless has no shell, so use ENTRYPOINT for health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD ["/app/api", "-healthcheck"]

# Application entry point
# Binary runs as non-root user (UID 65532) automatically
ENTRYPOINT ["/app/api"]
