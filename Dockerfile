# Use the official Go image as the base image
FROM golang:1.17-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy the necessary files to the container
COPY go.mod .
COPY go.sum .
COPY cmd/main.go ./cmd/main.go
COPY api/api.go ./api/api.go

# Build the Go application
RUN go build -o my-ipam-driver ./cmd/main.go

# Set the entrypoint command for the container
ENTRYPOINT ["./my-ipam-driver"]
