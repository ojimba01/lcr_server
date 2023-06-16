#!/bin/bash

echo "Building with Go..."
# Build the Docker image
cd backend

go mod download

go build -o backend

echo "Finishing up..."

./backend


echo "Hope Enjoyed LCR!"


# ----------------------------------------------

# To run this script, run the following command:

# chmod +x start.sh
# ./start.sh

# or

# sudo chmod +x start.sh
# sudo ./start.sh