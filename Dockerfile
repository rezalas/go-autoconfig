# Multi-stage build for minimal image size
FROM golang:1.26 AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o autoconfig

# Final stage - minimal runtime image
FROM alpine:latest

# Install ca-certificates for HTTPS and create non-root user
RUN apk --no-cache add ca-certificates && \
    addgroup -g 1000 autoconfig && \
    adduser -D -u 1000 -G autoconfig autoconfig

WORKDIR /app

# Copy binary and configs from builder
COPY --from=builder /build/autoconfig .
COPY --chown=autoconfig:autoconfig clientConfigs/ ./clientConfigs/
COPY --chown=autoconfig:autoconfig templates/ ./templates/

# Switch to non-root user
USER autoconfig

EXPOSE 8080

ENTRYPOINT ["/app/autoconfig"]
