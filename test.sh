#!/bin/bash -e
./run.sh
echo "Expecting 'Hello user!'"
curl -v -X GET http://localhost:8080/hello

echo "Expecting 'Hello Pete!'"
curl -v -X GET http://localhost:8080/hello?name=Pete

echo "Expecting 404"
curl -v -X PUT http://localhost:8080/hello

echo "Expecting 404"
curl -v -X PUT http://localhost:8080/hello?name=Pete

echo "Expecting 'This is a GET request'"
curl -v -X GET http://localhost:8080/check

echo "Expecting 'This is a POST request"
curl -v -X POST http://localhost:8080/check

echo "Expecting 405 Unsupported method"
curl -v -X PUT http://localhost:8080/check

echo "Tests completed.................."
./stop.sh
