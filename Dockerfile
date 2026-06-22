# ---- Stage 1: Build ----
# Use the full Go image to compile the binary
FROM golang:1.26-alpine AS builder

# Install git (some Go modules need it for fetching)
RUN apk add --no-cache git

WORKDIR /app

# Copy dependency files first (Docker caches this layer)
# If go.mod/go.sum haven't changed, Docker skips re-downloading deps
COPY go.mod go.sum ./
RUN go mod download

# Now copy the rest of the source code
COPY . .

# Build the binary
# CGO_ENABLED=0 : pure Go binary, no C dependencies (needed for Alpine)
# GOOS=linux    : target Linux (the container OS)
# -o webhook-server : output binary name
RUN CGO_ENABLED=0 GOOS=linux go build -o webhook-server .

# ---- Stage 2: Runtime ----
# Use a minimal Alpine image (5MB vs 800MB for full Go image)
FROM alpine:3.21

# Install CA certificates so the app can make HTTPS requests to webhook URLs
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy ONLY the compiled binary from the build stage
COPY --from=builder /app/webhook-server .

# Create an empty .env file so godotenv.Load(".env") doesn't crash
# The actual env vars are injected by Docker Compose into the OS environment
RUN touch .env

EXPOSE 3000

CMD ["./webhook-server"]