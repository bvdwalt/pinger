# Multi-stage build for pinger
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pinger .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/pinger .

# Copy config file - use config-example.yaml as default
COPY config.yaml .

# Create a non-root user
RUN adduser -D -u 1000 pinger && chown -R pinger:pinger /app
USER pinger

ENTRYPOINT ["./pinger"]
