# Stage 1: Build the Go binary
FROM golang:1.24 AS builder

WORKDIR /app

# Copy go.mod and go.sum first for dependency resolution
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN go build -o main .

# Stage 2: Runtime
FROM debian:bookworm-slim

# Install required packages (yt-dlp and ca-certificates for HTTPS requests)
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    curl ca-certificates ffmpeg python3 && \
    curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
    chmod a+rx /usr/local/bin/yt-dlp && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the compiled binary and any necessary files
COPY --from=builder /app/main /app/main
COPY .env ./

# Create media directory if not already in the build context
RUN mkdir -p media

# Command to run
CMD ["./main"]

# -- Build --
# docker build -t bot .
# 
# -- Run --
# docker run --env-file .env bot