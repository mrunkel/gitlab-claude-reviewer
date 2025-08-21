# Multi-stage build for smaller final image
FROM golang:1.21-alpine AS builder

# Install ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./
COPY go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gitlab-claude-reviewer ./cmd/gitlab-claude-reviewer

# Final stage - minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/gitlab-claude-reviewer .

# Make binary executable
RUN chmod +x ./gitlab-claude-reviewer

# Create non-root user
RUN adduser -D -s /bin/sh reviewer
USER reviewer

# Set entrypoint
ENTRYPOINT ["./gitlab-claude-reviewer"]