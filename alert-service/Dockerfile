# Build stage
FROM golang:1.25.4-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy all source code
COPY . .

# Build the application from cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o alert-service ./cmd/main.go

# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/alert-service .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./alert-service"]