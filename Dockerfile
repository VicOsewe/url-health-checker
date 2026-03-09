# Stage 1: Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency files first (better layer caching)
COPY go.mod ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o url-health-checker .

# Stage 2: Final minimal image
FROM alpine:3.19

WORKDIR /app

RUN apk --no-cache add ca-certificates

# create a non-root user
RUN adduser -D -g '' appuser

COPY --from=builder /app/url-health-checker .

# switch to non-root before running
USER appuser

ENTRYPOINT ["./url-health-checker"]