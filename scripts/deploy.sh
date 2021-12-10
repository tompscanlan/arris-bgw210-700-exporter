#!/bin/bash -x

go build -o arris-cli-darwin .
# GOOS=linux GOARCH=amd64 go build -o arris-cli-linux .
# scp arris-cli-linux admin@192.168.1.6:/bin/arris-cli

docker build -t tompscanlan/arris-cli-linux .
docker push tompscanlan/arris-cli-linux:latest