#!/bin/bash
echo "Executing run.sh..."
./run.sh

echo "Executing test_HW1.py..."
python test_HW1.py

echo "Tests completed.................."
./stop.sh
