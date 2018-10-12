#!/bin/bash
source values

echo "=========================>   RUNNING UNIT TESTS   <========================"
echo "$(go test -bench=. -v -coverprofile out && rm out)"

echo "=========================>  SETTING UP CONTAINER  <========================"
./run.sh

echo "=========================>   INTEGRATION TESTING  <========================"
${TEST_SCRIPT}

echo "=========================>       TEARDOWN         <========================"
./stop.sh
