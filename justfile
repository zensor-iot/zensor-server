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
    @if ! command -v mockgen &> /dev/null; then \
        echo "   - mockgen not found, installing..."; \
        go install go.uber.org/mock/mockgen@latest; \
    else \
        echo "   - mockgen already installed."; \
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
        docker compose up -d --wait
    fi
    echo "🚀 starting zensor server with hot reload..."
    find . -type f -name '*.go' | entr ./server

dev: build
    #!/bin/bash
    if [ "${ENV}" = "local" ]; then
        echo "🌱 Local mode: skipping Docker dependencies"
    else
        docker compose up -d --wait
    fi

    LOG_FILE=$(mktemp -t zensor-dev-XXXX.log)
    echo "🚀 starting zensor server (log: ${LOG_FILE})..."
    ./server > "$LOG_FILE" 2>&1 &
    SERVER_PID=$!
    TAIL_PID=""

    teardown() {
        echo ""
        echo "🔪 stopping server (PID: $SERVER_PID)..."
        [ -n "$TAIL_PID" ] && kill "$TAIL_PID" 2>/dev/null
        kill "$SERVER_PID" 2>/dev/null
        for _ in 1 2 3 4 5; do
            kill -0 "$SERVER_PID" 2>/dev/null || break
            sleep 1
        done
        kill -9 "$SERVER_PID" 2>/dev/null
        wait "$SERVER_PID" 2>/dev/null
    }
    trap teardown EXIT INT TERM

    echo "⏳ waiting for server to be ready..."
    max_attempts=30
    attempt=0
    while ! curl -sf http://127.0.0.1:3000/healthz > /dev/null; do
        if [ $attempt -ge $max_attempts ]; then
            echo "❌ server failed to start after 30 seconds."
            cat "$LOG_FILE"
            exit 1
        fi
        sleep 1
        attempt=$((attempt+1))
    done
    echo "✅ server is ready"

    ./scripts/seed.sh

    echo ""
    echo "📜 zensor server is running — tailing logs, press Ctrl+C to stop"
    tail -f "$LOG_FILE" &
    TAIL_PID=$!
    wait "$SERVER_PID"

seed:
    #!/bin/bash
    echo "🌱 seeding against an already-running server..."
    ./scripts/seed.sh

health:
    #!/bin/bash
    echo "🔍 checking service health..."

    # Check Postgres
    if nc -z localhost 5432; then
        echo "✅ postgresql: healthy (port 5432)"
    else
        echo "❌ postgresql: not responding on port 5432"
    fi

    # Check Redis
    if nc -z localhost 6379; then
        echo "✅ redis: healthy (port 6379)"
    else
        echo "❌ redis: not responding on port 6379"
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

mock: install-mockgen
    @echo "🔧 Generating mocks with comments..."
    @go generate ./internal/...
    @echo "✅ Mocks generated successfully!"

install-mockgen:
    @if ! command -v mockgen &> /dev/null; then \
        echo "📦 Installing mockgen..."; \
        go install go.uber.org/mock/mockgen@latest; \
    fi

mock-clean:
    @echo "🧹 Cleaning generated mocks..."
    @find . -name "*_mock.go" -type f -delete
    @echo "✅ Mocks cleaned!"

mock-interface interface path="internal":
    @echo "🔧 Generating mock for interface: {{interface}}"
    @mockgen -source={{path}} -destination={{path}}_mock.go -package=$(basename {{path}}) -mock_names={{interface}}=Mock{{interface}}

lint:
    golangci-lint run --max-issues-per-linter=0 --max-same-issues=0 --config=./build/ci/golangci.yml --timeout 7m

arch args="":
    arch-go {{args}}

tdd path="internal":
    go run github.com/onsi/ginkgo/v2/ginkgo watch --race {{path}}

unit path="internal":
    go run github.com/onsi/ginkgo/v2/ginkgo run -r --randomize-all --randomize-suites --fail-on-pending --keep-going --cover --coverprofile=coverprofile.out --race --trace --timeout=4m {{path}}

functional module tags="~@pending": build
    #!/bin/bash
    if [ -z "{{module}}" ]; then
        echo "❌ Module name is required. Usage: just functional <module> [tags]"
        echo "   Available modules: maintenance, permaculture"
        exit 1
    fi
    
    MODULE_PATH="test/functional/{{module}}"
    if [ ! -d "$MODULE_PATH" ]; then
        echo "❌ Module '{{module}}' not found at $MODULE_PATH"
        exit 1
    fi
    
    echo "🚀 Starting server in background..."
    export ENV=local
    export ZENSOR_SERVER_GENERAL_LOG_LEVEL=debug
    ./server > api.log 2>&1 &
    export SERVER_PID=$!
    
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
    while ! curl -sf http://127.0.0.1:3000/healthz > /dev/null; do
        if [ $attempt -ge $max_attempts ]; then
            echo "❌ Server failed to start after 30 seconds."
            echo "📋 Server log (api.log):"
            cat api.log 2>/dev/null || true
            exit 1
        fi
        sleep 1
        attempt=$((attempt+1))
    done
    
    echo "⏳ Verifying server stability (2s)..."
    sleep 2
    if ! curl -sf http://127.0.0.1:3000/healthz > /dev/null; then
        echo "❌ Server crashed during startup."
        echo "📋 Server log (api.log):"
        cat api.log 2>/dev/null || true
        exit 1
    fi
    echo "✅ Server is ready."
    
    echo "🧪 Running functional tests for module: {{module}}"
    echo "   - Running tests with tags: {{tags}}"
    go test ./$MODULE_PATH -v --godog.tags={{tags}}
    TEST_EXIT_CODE=$?
    
    if [ $TEST_EXIT_CODE -ne 0 ]; then
        echo "📋 Server log (api.log) for debugging:"
        cat api.log 2>/dev/null || true
    fi
    
    exit $TEST_EXIT_CODE

functional-module module tags="~@pending": build
    #!/bin/bash
    if [ -z "{{module}}" ]; then
        echo "❌ Module name is required. Usage: just functional-module <module> [tags]"
        echo "   Available modules: maintenance, permaculture"
        exit 1
    fi
    
    MODULE_PATH="test/functional/{{module}}"
    if [ ! -d "$MODULE_PATH" ]; then
        echo "❌ Module '{{module}}' not found at $MODULE_PATH"
        exit 1
    fi
    
    echo "🚀 Starting server in background..."
    export ENV=local
    export ZENSOR_SERVER_GENERAL_LOG_LEVEL=debug
    ./server > api.log 2>&1 &
    export SERVER_PID=$!
    
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
    while ! curl -sf http://127.0.0.1:3000/healthz > /dev/null; do
        if [ $attempt -ge $max_attempts ]; then
            echo "❌ Server failed to start after 30 seconds."
            echo "📋 Server log (api.log):"
            cat api.log 2>/dev/null || true
            exit 1
        fi
        sleep 1
        attempt=$((attempt+1))
    done
    
    echo "⏳ Verifying server stability (2s)..."
    sleep 2
    if ! curl -sf http://127.0.0.1:3000/healthz > /dev/null; then
        echo "❌ Server crashed during startup."
        echo "📋 Server log (api.log):"
        cat api.log 2>/dev/null || true
        exit 1
    fi
    echo "✅ Server is ready."
    
    echo "🧪 Running functional tests for module: {{module}}"
    echo "   - Running tests with tags: {{tags}}"
    go test ./$MODULE_PATH -v --godog.tags={{tags}}
    TEST_EXIT_CODE=$?
    
    if [ $TEST_EXIT_CODE -ne 0 ]; then
        echo "📋 Server log (api.log) for debugging:"
        cat api.log 2>/dev/null || true
    fi
    
    exit $TEST_EXIT_CODE

functional-external tags="@beta" api_url="http://localhost:3000":
    #!/bin/bash
    echo "🌍 Running functional tests against external API..."
    
    if [ -z "{{api_url}}" ]; then
        echo "❌ EXTERNAL_API_URL environment variable is required"
        exit 1
    fi
    
    echo "🔗 Target API URL: {{api_url}}"
    echo "🏷️  Running tests with tags: {{tags}}"
    
    cd test/functional
    EXTERNAL_API_URL="{{api_url}}" go test -v --godog.tags={{tags}}
    TEST_EXIT_CODE=$?
    
    if [ $TEST_EXIT_CODE -eq 0 ]; then
        echo "✅ External tests passed"
    else
        echo "❌ External tests failed"
    fi
    
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