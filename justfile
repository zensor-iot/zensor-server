default:
    @just --list

setup:
    #!/bin/bash
    echo "🚀 launching redpanda..."
    docker start redpanda || docker run --name redpanda -d -p 9092:9092 vectorized/redpanda:v22.2.7
    while ! nc -z localhost 9092; do
        sleep 0.5
    done
    rpk topic --brokers "localhost:9092" describe device_registered || rpk topic --brokers "localhost:9092" create device_registered
    echo "🚀 launching materialize..."
    docker start materialize || docker run --name materialize -p 6875:6875 -d materialize/materialized:v0.7.3 --workers 1
    while ! nc -z localhost 9092; do
        sleep 0.5
    done
    echo "🚀 launching mystique..."
    docker start mystique || docker run --name mystique -d -p 1883:1883 thethingsindustries/mystique-server

run: build
    ./server

build:
    go build -o server cmd/api/main.go

docker-build: build
    docker build -t zensor/server .
