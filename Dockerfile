FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -ldflags "-X main.version=${VERSION}" -o pinger ./cmd/pinger

FROM alpine:latest
RUN apk --no-cache add ca-certificates && \
    addgroup -S pinger && adduser -S pinger -G pinger
WORKDIR /app
COPY --from=builder /app/pinger .
USER pinger
ENTRYPOINT ["./pinger", "-config", "/config/config.yaml"]
