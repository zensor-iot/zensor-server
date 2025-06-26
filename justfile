default:
    @just --list

init:
    just install-deps

install-deps: install-otelcol
    @echo "📦 Installing Go tools..."
    @if ! command -v golangci-lint &> /dev/null; then \
        echo "   - golangci-lint not found, installing..."; \
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
    else \
        echo "   - golangci-lint already installed."; \
    fi
    @if ! command -v ginkgo &> /dev/null; then \
        echo "   - ginkgo not found, installing..."; \
        go install github.com/onsi/ginkgo/v2/ginkgo@latest; \
    else \
        echo "   - ginkgo already installed."; \
    fi
    @if ! command -v wire &> /dev/null; then \
        echo "   - wire not found, installing..."; \
        go install github.com/google/wire/cmd/wire@latest; \
    else \
        echo "   - wire already installed."; \
    fi
    @if ! command -v arch-go &> /dev/null; then \
        echo "   - arch-go not found, installing..."; \
        go install github.com/arch-go/arch-go@latest; \
    else \
        echo "   - arch-go already installed."; \
    fi
    @echo "✅ All dependencies installed."

install-otelcol:
    #!/bin/bash
    if [ -f "./otelcol" ]; then
        echo "otelcol is already installed."
        exit 0
    fi
    arch=$(uname -m)
    if [[ "$arch" == "x86_64" ]]; then
        arch="amd64"
    fi
    echo "Installing otelcol for darwin/${arch}..."
    mkdir -p tmp
    pushd tmp
    curl --proto '=https' --tlsv1.2 -fOL "https://github.com/open-telemetry/opentelemetry-collector-releases/releases/download/v0.108.0/otelcol_0.108.0_darwin_${arch}.tar.gz"
    tar -xvf "otelcol_0.108.0_darwin_${arch}.tar.gz"
    popd
    mv tmp/otelcol .
    rm -rf tmp

setup:
    #!/bin/bash
    echo "🚀 creating network..."
    docker network inspect zensor || docker network create zensor
    
    echo "🚀 launching redpanda..."
    docker start redpanda || docker container run --name redpanda --network zensor -d -p 19092:19092 redpandadata/redpanda:v24.3.4 redpanda start --kafka-addr internal://0.0.0.0:9092,external://0.0.0.0:19092  --advertise-kafka-addr internal://redpanda:9092,external://localhost:19092
    echo "⏳ waiting for redpanda to be ready..."
    while ! nc -z localhost 19092; do
        sleep 0.5
    done
    echo "✅ redpanda is ready"
    
    echo "🔧 creating kafka topics..."
    rpk topic --brokers "localhost:19092" describe devices || rpk topic --brokers "localhost:19092" create devices
    rpk topic --brokers "localhost:19092" describe evaluation_rules || rpk topic --brokers "localhost:19092" create evaluation_rules
    rpk topic --brokers "localhost:19092" describe device_commands || rpk topic --brokers "localhost:19092" create device_commands
    rpk topic --brokers "localhost:19092" describe tasks || rpk topic --brokers "localhost:19092" create tasks
    rpk topic --brokers "localhost:19092" describe tenants || rpk topic --brokers "localhost:19092" create tenants
    rpk topic --brokers "localhost:19092" describe scheduled_tasks || rpk topic --brokers "localhost:19092" create scheduled_tasks
    echo "✅ kafka topics created"
    
    echo "🚀 launching materialize..."
    docker start materialize || docker container run --name materialize --network zensor -p 6875:6875 -d materialize/materialized:v0.133.0-dev.0--main.gd098b5f47028a4eccd4b3bc4ce6f8cd33c1895cf
    echo "⏳ waiting for materialize port to be available..."
    
    # Wait for port to be available
    while ! nc -z localhost 6875; do
        sleep 1
    done
    echo "✅ materialize port is available (full validation will happen before starting app)"
    
    echo "🚀 launching prometheus..."
    docker start prometheus || docker container run --name prometheus --network zensor -p 9090:9090 -d bitnami/prometheus:2.55.1 --config.file=/opt/bitnami/prometheus/conf/prometheus.yml --storage.tsdb.path=/opt/bitnami/prometheus/data --web.console.libraries=/opt/bitnami/prometheus/conf/console_libraries --web.console.templates=/opt/bitnami/prometheus/conf/consoles --web.enable-remote-write-receiver
    echo "⏳ waiting for prometheus to be ready..."
    while ! nc -z localhost 9090; do
        sleep 0.5
    done
    echo "✅ prometheus is ready"
    
    echo "🚀 launching grafana..."
    docker start grafana || docker container run --name grafana --network zensor -p 3001:3000 -d grafana/grafana:11.5.1
    echo "✅ grafana started"
    
    echo "🎉 all services are ready!"

build:
    go build -o server cmd/api/main.go

run: build
    #!/bin/bash
    if [ "${ENV}" = "local" ]; then
        echo "🌱 Local mode: skipping Docker dependencies (setup/validate-db)..."
    else
        just setup
        just validate-db
    fi
    echo "🔧 starting opentelemetry collector..."
    ./otelcol --config otelcol_config.yaml > otelcol.log 2>&1 &
    echo "🚀 starting zensor server with hot reload..."
    find . -type f -name '*.go' | entr ./server

validate-db:
    #!/bin/bash
    echo "🔍 validating materialize database connectivity..."
    
    max_attempts=60
    attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if psql -h localhost -p 6875 -U materialize -d materialize -c "SELECT 1;" >/dev/null 2>&1; then
            echo "✅ materialize database is ready and accepting connections"
            
            # Additional validation: check if we can create a simple test connection
            if psql -h localhost -p 6875 -U materialize -d materialize -c "SELECT version();" >/dev/null 2>&1; then
                echo "✅ materialize database validation successful"
                exit 0
            else
                echo "⚠️  materialize responds but may not be fully ready"
            fi
        fi
        
        echo "⏳ attempt $attempt/$max_attempts: waiting for materialize database..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "❌ materialize database failed to become ready after $max_attempts attempts ($(($max_attempts * 2)) seconds)"
    echo "🔍 checking materialize container status..."
    docker ps --filter "name=materialize"
    echo "🔍 checking materialize logs..."
    docker logs materialize --tail 20
    exit 1

migrate:
    #!/bin/bash
    echo "🔄 applying database migrations..."
    
    # Apply migrations by reading all .sql files in order
    for migration in migrations/*.up.sql; do
        if [ -f "$migration" ]; then
            echo "📄 applying $(basename "$migration")..."
            if ! psql -h localhost -p 6875 -U materialize -d materialize -f "$migration"; then
                echo "❌ failed to apply $(basename "$migration")"
                exit 1
            fi
        fi
    done
    echo "✅ migrations completed successfully"

health:
    #!/bin/bash
    echo "🔍 checking service health..."
    
    # Check Redpanda
    if nc -z localhost 19092; then
        echo "✅ redpanda: healthy (port 19092)"
    else
        echo "❌ redpanda: not responding on port 19092"
    fi
    
    # Check Materialize
    if psql -h localhost -p 6875 -U materialize -d materialize -c "SELECT 1;" >/dev/null 2>&1; then
        echo "✅ materialize: healthy (port 6875)"
    else
        echo "❌ materialize: not responding on port 6875"
    fi
    
    # Check Prometheus
    if nc -z localhost 9090; then
        echo "✅ prometheus: healthy (port 9090)"
    else
        echo "❌ prometheus: not responding on port 9090"
    fi
    
    # Check Grafana
    if nc -z localhost 3001; then
        echo "✅ grafana: healthy (port 3001)"
    else
        echo "❌ grafana: not responding on port 3001"
    fi

destroy:
    #!/bin/bash
    echo "🧹 stopping and removing containers..."
    docker stop redpanda | xargs docker rm
    docker stop materialize | xargs docker rm
    docker stop prometheus | xargs docker rm
    docker stop grafana | xargs docker rm
    echo "✅ cleanup completed"


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

functional tags="": build setup validate-db
    #!/bin/bash
    echo "🚀 Starting server in background..."
    export ENV=local
    export ZENSOR_SERVER_GENERAL_LOG_LEVEL=debug
    ./server > api.log 2>&1 &
    SERVER_PID=$!
    
    # Teardown function to ensure the server is killed
    teardown() {
        echo "🔪 Tearing down server (PID: $SERVER_PID)..."
        kill $SERVER_PID
        wait $SERVER_PID 2>/dev/null
    }
    
    # Trap exit signals to ensure teardown runs
    trap teardown EXIT
    
    echo "⏳ Waiting for server to be ready..."
    max_attempts=30
    attempt=0
    while ! nc -z localhost 3000; do
        if [ $attempt -ge $max_attempts ]; then
            echo "❌ Server failed to start after 30 seconds."
            exit 1
        fi
        sleep 1
        attempt=$((attempt+1))
    done
    echo "✅ Server is ready."
    
    echo "🧪 Running functional tests..."
    echo "   - Running tests with tags: {{tags}}"
    cd test/functional
    go test -v --godog.tags={{tags}}
    TEST_EXIT_CODE=$?
    
    exit $TEST_EXIT_CODE

c4:
    docker run -it \
        --rm \
        -p 8080:8080 \
        -v "$(pwd)/docs":/usr/local/structurizr \
        -e STRUCTURIZR_WORKSPACE_PATH=. \
        -e STRUCTURIZR_WORKSPACE_FILENAME=c4model \
        structurizr/lite

release version:
    #!/bin/bash
    git tag {{version}}
    git push --tags