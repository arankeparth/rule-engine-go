# Use the official Golang image as the base image
FROM golang:1.19-alpine AS builder

# Set the working directory in the container
WORKDIR /app

# Copy the Go modules files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire application code to the container
COPY . .

# Build the Go app
RUN go build -o server .

# Use a minimal base image for the final stage
FROM alpine:latest

# Set the working directory in the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/server .

# Copy any additional files needed at runtime, such as configuration files
COPY rules.json .

COPY responses ./responses

# Expose the port on which the server will run
EXPOSE 8081

# Run the binary
CMD ["./server"]
