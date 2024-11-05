#!/usr/bin/env bash

set -e

# Define the name of the Docker image
IMAGE_NAME=setddblock-builder

# Build the Docker image for macOS
docker build --build-arg GOOS=darwin --build-arg GOARCH=arm64 -t ${IMAGE_NAME}-macos .

# Create a container from the macOS image
CONTAINER_ID_MACOS=$(docker create ${IMAGE_NAME}-macos)

# Copy the macOS binary from the container to the host
docker cp $CONTAINER_ID_MACOS:/setddblock ./setddblock-macos

# Remove the macOS container
docker rm $CONTAINER_ID_MACOS

# Build the Docker image for Linux
docker build --build-arg GOOS=linux --build-arg GOARCH=amd64 -t ${IMAGE_NAME}-linux .

# Create a container from the Linux image
CONTAINER_ID_LINUX=$(docker create ${IMAGE_NAME}-linux)

# Copy the Linux binary from the container to the host
docker cp $CONTAINER_ID_LINUX:/setddblock ./setddblock-linux

# Remove the Linux container
docker rm $CONTAINER_ID_LINUX

# Create a container from the image
CONTAINER_ID=$(docker create $IMAGE_NAME)

# Copy the binary from the container to the host
docker cp $CONTAINER_ID:/setddblock ./setddblock

# Remove the container
docker rm $CONTAINER_ID

echo "Build complete. The binary is located at ./setddblock"
