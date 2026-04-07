# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/mystisql ./cmd/mystisql

# Build frontend
FROM node:25-alpine AS frontend-builder

WORKDIR /web

# Copy package files
COPY web/package.json web/package-lock.json* ./

# Install dependencies
RUN npm ci || npm install

# Copy frontend source
COPY web/ .

# Build frontend
RUN npm run build

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/mystisql /app/mystisql

# Copy frontend dist from frontend-builder
COPY --from=frontend-builder /web/dist /app/web/dist

# Copy default config (optional)
COPY config.yaml /app/config.yaml

# Create log directory
RUN mkdir -p /var/log/mystisql

# Expose port
EXPOSE 8080

# Set environment
ENV GIN_MODE=release

# Run the binary
ENTRYPOINT ["/app/mystisql"]
CMD ["serve"]
