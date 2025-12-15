# Frontend build stage
FROM --platform=$BUILDPLATFORM  node:24 AS frontend-builder

WORKDIR /app
COPY frontend ./frontend
WORKDIR /app/frontend
RUN npm install -g pnpm && pnpm install && pnpm build

# Build stage
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS builder

WORKDIR /app

# Install git (needed for fetching dependencies)
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -a -installsuffix cgo -o github-stars-manager .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/github-stars-manager .

# Create directories
RUN mkdir -p data static

# Copy templates
COPY --from=frontend-builder /app/frontend/dist/*.html templates/
# Copy static assets
COPY --from=frontend-builder /app/frontend/dist/static/ static/

# Expose port
EXPOSE 8181

# Environment variables with default values
ENV GITHUB_CLIENT_ID=你的ClientID
ENV GITHUB_CLIENT_SECRET=你的ClientSecret
ENV GITHUB_REDIRECT_URL=http://localhost:8181/auth/github/callback
ENV SERVER_PORT=:8181
ENV LOGGER_LEVEL=info

# Command to run the executable
CMD ["./github-stars-manager"]