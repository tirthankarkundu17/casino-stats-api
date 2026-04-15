#!/bin/bash

# Script to run the seeder one-time using Docker Compose
echo "Starting data seeding process..."

# Run the seeder service. 
# --rm removes the container after it finishes.
docker-compose run --rm seeder

echo "Seeding completed successfully!"
