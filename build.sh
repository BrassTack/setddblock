#!/usr/bin/env bash

set -e

# Define the name of the Docker image
IMAGE_NAME=setddblock-builder

echo "Starting Docker build for macOS ARM64..."

# Build the Docker image for macOS ARM64 with no cache, specifying the Dockerfile explicitly
docker build --no-cache --progress=plain --file Dockerfile --build-arg GOOS=darwin --build-arg GOARCH=arm64 --build-arg BUILD_DATE=$(date +%Y-%m-%dT%H:%M:%S) -t ${IMAGE_NAME}-macos-arm64 .

echo "Docker build for macOS ARM64 completed."

# Create a container from the macOS ARM64 image
echo "Creating container from the macOS ARM64 image..."
CONTAINER_ID_MACOS_ARM64=$(docker create ${IMAGE_NAME}-macos-arm64)

# Copy the macOS ARM64 binary from the container to the host
echo "Copying macOS ARM64 binary from the container to the host..."
docker cp $CONTAINER_ID_MACOS_ARM64:/setddblock ./setddblock-macos-arm64

# Remove the macOS ARM64 container
echo "Removing the macOS ARM64 container..."
docker rm $CONTAINER_ID_MACOS_ARM64

echo "Build complete. The binary is located at ./setddblock-macos-arm64"

echo "Starting Docker build for Linux AMD64..."

# Build the Docker image for Linux AMD64 with no cache, specifying the Dockerfile explicitly
docker build --no-cache --progress=plain --file Dockerfile --build-arg GOOS=linux --build-arg GOARCH=amd64 --build-arg BUILD_DATE=$(date +%Y-%m-%dT%H:%M:%S) -t ${IMAGE_NAME}-linux-amd64 .

echo "Docker build for Linux AMD64 completed."

# Create a container from the Linux AMD64 image
echo "Creating container from the Linux AMD64 image..."
CONTAINER_ID_LINUX_AMD64=$(docker create ${IMAGE_NAME}-linux-amd64)

# Copy the Linux AMD64 binary from the container to the host
echo "Copying Linux AMD64 binary from the container to the host..."
docker cp $CONTAINER_ID_LINUX_AMD64:/setddblock ./setddblock-linux-amd64

# Remove the Linux AMD64 container
echo "Removing the Linux AMD64 container..."
docker rm $CONTAINER_ID_LINUX_AMD64

echo "Build complete. The binary is located at ./setddblock-linux-amd64"
