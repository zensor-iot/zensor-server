default:
    @just --list

setup:
    #!/bin/bash
    echo "🚀 creating network..."
    docker network inspect zensor || docker network create zensor
    echo "🚀 launching redpanda..."
    docker start redpanda || docker container run --name redpanda --network zensor -d -p 19092:19092 redpandadata/redpanda:v24.3.4 redpanda start --kafka-addr internal://0.0.0.0:9092,external://0.0.0.0:19092  --advertise-kafka-addr internal://redpanda:9092,external://localhost:19092
    while ! nc -z localhost 19092; do
        sleep 0.5
    done
    rpk topic --brokers "localhost:19092" describe devices || rpk topic --brokers "localhost:19092" create devices
    echo "🚀 launching materialize..."
    docker start materialize || docker container run --name materialize --network zensor -p 6875:6875 -d materialize/materialized:v0.133.0-dev.0--main.gd098b5f47028a4eccd4b3bc4ce6f8cd33c1895cf
    while ! nc -z localhost 6875; do
        sleep 0.5
    done

run: build
    ./server

build:
    go build -o server cmd/api/main.go

docker-build: build
    docker build -t zensor/server .
