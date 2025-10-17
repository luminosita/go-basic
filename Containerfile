# =============================================================================
# Multi-Stage Production Containerfile for CHANGE_ME
# =============================================================================
# Compatible with: Podman 4.0+, Docker 20.10+
# Target image size: <500MB
# Security: Non-root user execution (UID 1000)
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Builder - Install dependencies and build artifacts
# -----------------------------------------------------------------------------
FROM python:3.11-slim AS builder

# Install build dependencies for compiling Python packages
# gcc: C compiler for building Python C extensions
# build-essential: Essential build tools (make, etc.)
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    gcc=4:12.2.0-3 \
    build-essential=12.9 \
    curl=7.88.1-10+deb12u8 \
    && rm -rf /var/lib/apt/lists/*

# Install uv - fast Python package installer
# WHY UV: Significantly faster than pip for dependency resolution and installation
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN curl -LsSf https://astral.sh/uv/install.sh | sh
ENV PATH="/root/.local/bin:${PATH}"

# Set working directory for build
WORKDIR /build

# Copy dependency files and source code
# Note: Source code needed for hatchling to build package during uv sync
COPY pyproject.toml uv.lock README.md ./
COPY src/ ./src/

# Install Python dependencies using uv
# --frozen: Use exact versions from uv.lock (deterministic builds)
# --no-dev: Skip development dependencies (tests, linters)
RUN uv sync --frozen --no-dev

# -----------------------------------------------------------------------------
# Stage 2: Production - Minimal runtime image
# -----------------------------------------------------------------------------
FROM python:3.11-slim AS production

# Install runtime dependencies only
# libpq5: PostgreSQL client library (required by asyncpg)
# ca-certificates: SSL/TLS certificate validation
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    libpq5=15.10-0+deb12u1 \
    ca-certificates=20230311 \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user for security
# appuser (UID 1000): Standard non-privileged user
# WHY: Principle of least privilege - reduces attack surface if container compromised
RUN groupadd -g 1000 appuser && \
    useradd -u 1000 -g appuser -m -s /bin/bash appuser

# Set working directory
WORKDIR /app

# Copy Python dependencies from builder stage
# Only virtual environment, not build tools
COPY --from=builder --chown=appuser:appuser /build/.venv /app/.venv

# Copy application source code from builder stage
COPY --from=builder --chown=appuser:appuser /build/src /app/src

# Set Python path to use virtual environment
ENV PATH="/app/.venv/bin:${PATH}" \
    PYTHONPATH="/app/src" \
    PYTHONUNBUFFERED=1

# Switch to non-root user
# All subsequent commands and container runtime execute as appuser
USER appuser

# Expose application port
# Port 8000: FastAPI default port (Uvicorn)
EXPOSE 8000

# Health check configuration
# Interval: Check every 30 seconds
# Timeout: Fail if health check takes >10 seconds
# Start period: Wait 40 seconds after container start before first check
# Retries: Mark unhealthy after 3 consecutive failures
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8000/health').read()" || exit 1

# Application entry point
# Uvicorn: ASGI server for FastAPI
# --host 0.0.0.0: Listen on all interfaces (required for container networking)
# --port 8000: Standard FastAPI port
# Using python -m syntax for better compatibility
CMD ["python", "-m", "uvicorn", "change_me.main:app", "--host", "0.0.0.0", "--port", "8000"]
