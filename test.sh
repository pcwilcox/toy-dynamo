#!/bin/bash
source values

echo "=========================>   RUNNING UNIT TESTS   <========================"
echo "$(make test)"
echo "$(make cover)"
echo "$(make bench)"

echo "=========================>  SETTING UP CONTAINER  <========================"
./run.sh

echo "=========================>   INTEGRATION TESTING  <========================"
${TEST_SCRIPT}

echo "=========================>       TEARDOWN         <========================"
./stop.sh
