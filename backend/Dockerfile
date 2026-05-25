# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o heka ./cmd/server

# Final stage - use distroless for minimal size
FROM gcr.io/distroless/static-debian12:nonroot

# Install CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder
COPY --from=builder --chown=nonroot:nonroot /build/heka /heka

# Create directories for uploads
RUN mkdir -p /uploads && chown -R nonroot:nonroot /uploads

# Set user to nonroot
USER nonroot

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/heka", "--healthcheck"]

# Run the application
ENTRYPOINT ["/heka"]
