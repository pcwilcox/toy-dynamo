#!/bin/bash -e
source values
echo "=========================>   BUILDING DOCKERFILE  <========================"
docker build ${BUILD_FLAGS} .

echo "=========================>   RUNNING CONTAINER    <========================"
docker run ${RUN_FLAGS}

echo "=========================>      APP RUNNING       <========================"
echo ""
echo "App is now listening on: http://localhost:${PORT_EXT}"
echo ""
echo "To attach to console output:"
echo ""
echo "  docker attach ${NAME}"
echo ""
echo "To terminate app and remove container:"
echo "  stop.sh"
echo ""