#!/bin/bash -e


for key in $(cat ./keys_small.txt); do curl -X PUT http://localhost:8082/keyValue-store/${key} -d "val=hello"; done