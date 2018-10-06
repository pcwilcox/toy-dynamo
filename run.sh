#!/bin/bash -e
echo "Building new image.............."
docker image build --tag local:cs128_hw1 .

echo "Starting app...................."
docker run -d -p 8080:8080 --name cs128_hw1 local:cs128_hw1
