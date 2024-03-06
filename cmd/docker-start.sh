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

# ----------------------------------------------

# To run this script, run the following command:

# chmod +x local-start.sh
# ./start-start.sh

# or

# sudo chmod +x docker-start.sh
# sudo ./docker-start.sh