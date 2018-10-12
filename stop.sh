#!/bin/bash -e
source values

echo "============> STOPPING CONTAINER <============"
docker stop ${NAME}

echo "============> REMOVING IMAGE <============"
docker image rm ${TAG}

echo "============> TEARDOWN COMPLETE <============"

