#!/bin/bash -e
echo "Stopping container ........."
docker stop cs128_hw1

echo "Removing container........."
docker container rm cs128_hw1

echo "All done..................."

