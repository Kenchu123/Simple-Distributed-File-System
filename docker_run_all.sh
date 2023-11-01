#!/bin/bash
for i in {1..10}
do
   docker exec -d dev-m$i go run main.go serve
done
