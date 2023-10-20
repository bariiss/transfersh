# Start from the latest Go base image
FROM golang:latest AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code from the current directory to the working directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o transfersh .

### Start a new stage from scratch ###
FROM alpine:latest

WORKDIR /root/

# Copy the pre-built binary from the previous stage
COPY --from=builder /app/transfersh-cli .

# Command to run the executable
ENTRYPOINT ["./transfersh-cli"]
