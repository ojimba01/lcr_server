#!/bin/bash

echo "Building the Docker image..."
# Prompt user for POSTGRES_PASSWORD
read -sp "Enter your POSTGRES_PASSWORD: " POSTGRES_PASSWORD
echo "Processing... "

cd ../

# Build the Docker image
docker build --build-arg POSTGRES_PASSWORD="$POSTGRES_PASSWORD" -t myapp .

echo "Starting the Docker container..."
# Run the Docker container
docker run -p 3000:3000 -e POSTGRES_PASSWORD="$POSTGRES_PASSWORD" myapp

echo "Hope you enjoyed LCR!"