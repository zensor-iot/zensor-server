default:
    @just --list

build:
    go build -o server cmd/api/main.go

docker-build: build
    docker build -t zensor/server .