# Start with the official Golang image as the build environment
FROM golang:1.21 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Build the Go app with specified OS and architecture
ARG GOOS
ARG GOARCH
ARG BUILD_DATE
ENV CGO_ENABLED=0

# Copy the source code into the container
COPY . .

RUN echo "Build date: $BUILD_DATE" && echo "Building for OS: $GOOS, ARCH: $GOARCH" && uname -a && file /usr/local/go/bin/go && echo "Go environment variables:" && go env && echo "Go version:" && go version && GOOS=$GOOS GOARCH=$GOARCH go build -v -a -installsuffix cgo -ldflags '-s -w' -o /setddblock cmd/setddblock/main.go && file /setddblock

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
RUN go build -v -o /setddblock cmd/setddblock/main.go

# Start a new stage from scratch
FROM alpine:latest

# Copy the pre-built binary file from the previous stage
COPY --from=builder /setddblock /setddblock

# Command to run the executable
ENTRYPOINT ["/setddblock"]
