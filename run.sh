#!/bin/bash -e
echo "Building new image.............."
docker image build --tag local:rest_api .

echo "Starting app...................."
docker run -d -p 8080:8080 --name rest_api local:rest_api
