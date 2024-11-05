#!/usr/bin/env bash

set -e

# Define the name of the Docker image
IMAGE_NAME=setddblock-builder

#!/usr/bin/env bash

set -e

# Define the name of the Docker image
IMAGE_NAME=setddblock-builder

# Build the Docker image for macOS ARM64
docker build --build-arg GOOS=darwin --build-arg GOARCH=arm64 --build-arg BUILD_DATE=$(date +%Y-%m-%dT%H:%M:%S) -t ${IMAGE_NAME}-macos-arm64 .

# Create a container from the macOS ARM64 image
CONTAINER_ID_MACOS_ARM64=$(docker create ${IMAGE_NAME}-macos-arm64)

# Copy the macOS ARM64 binary from the container to the host
docker cp $CONTAINER_ID_MACOS_ARM64:/setddblock ./setddblock-macos-arm64

# Remove the macOS ARM64 container
docker rm $CONTAINER_ID_MACOS_ARM64

# Build the Docker image for Linux AMD64
docker build --build-arg GOOS=linux --build-arg GOARCH=amd64 --build-arg BUILD_DATE=$(date +%Y-%m-%dT%H:%M:%S) -t ${IMAGE_NAME}-linux-amd64 .

# Create a container from the Linux AMD64 image
CONTAINER_ID_LINUX_AMD64=$(docker create ${IMAGE_NAME}-linux-amd64)

# Copy the Linux AMD64 binary from the container to the host
docker cp $CONTAINER_ID_LINUX_AMD64:/setddblock ./setddblock-linux-amd64

# Remove the Linux AMD64 container
docker rm $CONTAINER_ID_LINUX_AMD64

echo "Build complete. The binaries are located at ./setddblock-macos-arm64 and ./setddblock-linux-amd64"
