#!/bin/bash

echo "Building with Go..."
# Build the Docker image
cd ../backend

go mod download

go build -o backend

echo "Finishing up..."

./backend

echo "Hope Enjoyed LCR!"


# ----------------------------------------------

# To run this script, run the following command:

# chmod +x local-start.sh
# ./start-start.sh

# or

# sudo chmod +x local-start.sh
# sudo ./local-start.sh