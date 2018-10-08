#!/bin/bash -e
source values
echo "=========================>   BUILDING DOCKERFILE  <========================"
docker build ${BUILD_FLAGS} .

echo "=========================>   RUNNING CONTAINER    <========================"
docker run ${RUN_FLAGS}

echo "=========================>      APP RUNNING       <========================"
echo ""
echo "App is now running on: http://localhost:${PORT_EXT}"
echo ""
echo 'Execute "docker attach ${NAME}" to attach to console output.'
echo 'Execute "stop.sh" to terminate app and remove container.'