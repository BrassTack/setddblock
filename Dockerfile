# Start with the official Golang image as the build environment
FROM golang:1.21 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app with specified OS and architecture
ARG GOOS
ARG GOARCH
ENV CGO_ENABLED=0
RUN GOOS=$GOOS GOARCH=$GOARCH go build -o /setddblock cmd/setddblock/main.go

# Start a new stage from scratch
FROM alpine:latest

# Copy the pre-built binary file from the previous stage
COPY --from=builder /setddblock /setddblock

# Command to run the executable
ENTRYPOINT ["/setddblock"]
# Start with the official Golang image as the build environment
FROM golang:1.21 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o /setddblock cmd/setddblock/main.go

# Start a new stage from scratch
FROM alpine:latest

# Copy the pre-built binary file from the previous stage
COPY --from=builder /setddblock /setddblock

# Command to run the executable
ENTRYPOINT ["/setddblock"]
