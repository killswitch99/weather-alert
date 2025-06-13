FROM golang:1.23-alpine

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Set Go environment variables
ENV GO111MODULE=on
ENV CGO_ENABLED=0

# Copy go mod and sum files first
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Install air for hot reloading with specific version
RUN go install github.com/air-verse/air@v1.52.3

# Copy source code
COPY . .

# Expose port 8080
EXPOSE 8080

# Command to run the application with hot reloading
CMD ["air", "-c", ".air.toml"]
