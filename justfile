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

build:
    go build -o server cmd/api/main.go

run: build
    #!/bin/bash
    if [ "${ENV}" = "local" ]; then
        echo "🌱 Local mode: skipping Docker dependencies"
    else
        docker compose up -d
    fi
    echo "🔧 starting opentelemetry collector..."
    ./otelcol --config otelcol_config.yaml > otelcol.log 2>&1 &
    echo "🚀 starting zensor server with hot reload..."
    find . -type f -name '*.go' | entr ./server

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
    docker compose down


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
    go run github.com/onsi/ginkgo/v2/ginkgo watch --race {{path}}

unit path="internal":
    go run github.com/onsi/ginkgo/v2/ginkgo run -r --randomize-all --randomize-suites --fail-on-pending --keep-going --cover --coverprofile=coverprofile.out --race --trace --timeout=4m {{path}}

functional tags="~@pending": build
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