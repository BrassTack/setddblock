#!/usr/bin/env bash

set -e

# Define the name of the Docker image
IMAGE_NAME=setddblock-builder

# Build the Docker image
docker build -t $IMAGE_NAME .

# Create a container from the image
CONTAINER_ID=$(docker create $IMAGE_NAME)

# Copy the binary from the container to the host
docker cp $CONTAINER_ID:/setddblock ./setddblock

# Remove the container
docker rm $CONTAINER_ID

echo "Build complete. The binary is located at ./setddblock"
