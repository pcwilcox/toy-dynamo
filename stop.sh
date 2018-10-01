#!/bin/bash -e
echo "Stopping container ........."
docker stop rest_api

echo "Removing container........."
docker container rm rest_api

echo "All done..................."

