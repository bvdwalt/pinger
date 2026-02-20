# Multi-stage build for pinger
FROM golang:1.26-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the binary with optimization flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a \
    -installsuffix cgo \
    -ldflags="-s -w" \
    -trimpath \
    -o pinger ./cmd/pinger

# Final stage - use distroless for minimal size
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/pinger .

# Copy config file - use config.yaml as default
COPY config.yaml .

ENTRYPOINT ["/app/pinger"]
