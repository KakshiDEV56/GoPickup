# Start from the official Golang image for building
FROM golang:1.21.6-alpine3.19 AS builder

# Set environment variables
ENV GO111MODULE=on

# Set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o app .

# Start a new, minimal image for running
FROM alpine:3.19

# Install CA certificates
RUN apk --no-cache upgrade && apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/app .

# Copy any static/config files if needed (uncomment if required)
# COPY --from=builder /app/config ./config

# Expose the port your app runs on (change if needed)
EXPOSE 8070

# Command to run the executable
CMD ["./app"]