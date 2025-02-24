default:
    @just --list

init:
    just install_otelcol

install_otelcol:
    #!/bin/bash
    arch=$(uname -m)
    mkdir tmp
    pushd tmp
    curl --proto '=https' --tlsv1.2 -fOL https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.108.0/otelcol_0.108.0_darwin_${arch}.tar.gz
    tar -xvf otelcol_0.108.0_darwin_arm64.tar.gz
    popd
    mv tmp/otelcol .
    rm -rf tmp

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
    rpk topic --brokers "localhost:19092" describe device_commands || rpk topic --brokers "localhost:19092" create device_commands
    echo "🚀 launching materialize..."
    docker start materialize || docker container run --name materialize --network zensor -p 6875:6875 -d materialize/materialized:v0.133.0-dev.0--main.gd098b5f47028a4eccd4b3bc4ce6f8cd33c1895cf
    while ! nc -z localhost 6875; do
        sleep 0.5
    done
    echo "🚀 launching prometheus..."
    docker start prometheus || docker container run --name prometheus --network zensor -p 9090:9090 -d bitnami/prometheus:2.55.1 --config.file=/opt/bitnami/prometheus/conf/prometheus.yml --storage.tsdb.path=/opt/bitnami/prometheus/data --web.console.libraries=/opt/bitnami/prometheus/conf/console_libraries --web.console.templates=/opt/bitnami/prometheus/conf/consoles --web.enable-remote-write-receiver
    while ! nc -z localhost 9090; do
        sleep 0.5
    done
    docker start grafana || docker container run --name grafana --network zensor -p 3001:3000 -d grafana/grafana:11.5.1

run: build setup
    #!/bin/bash
    ./otelcol --config otelcol_config.yaml > otelcol.log 2>&1 &
    ./server

build:
    go build -o server cmd/api/main.go

docker-build: build
    docker build -t zensor/server .

wire:
  cd cmd/api/wire && wire

mock:
  go generate ./internal/...

lint:
  golangci-lint run --max-issues-per-linter=0 --max-same-issues=0 --config=./build/ci/golangci.yml --timeout 7m

arch args="":
    arch-go {{args}}

tdd path="internal":
  ginkgo watch --race {{path}}

unit path="internal":
  ginkgo run -r --randomize-all --randomize-suites --fail-on-pending --keep-going --cover --coverprofile=coverprofile.out --race --trace --timeout=4m {{path}}
