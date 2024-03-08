default:
    @just --list

run: build
    ./server

build:
    go build -o server cmd/api/main.go

docker-build: build
    docker build -t zensor/server .