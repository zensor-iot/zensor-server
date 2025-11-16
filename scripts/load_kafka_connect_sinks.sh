#!/bin/bash

# Script to load all Kafka Connect sink connectors from JSON files
# Usage: ./scripts/load_kafka_connect_sinks.sh [kafka-connect-url]

KAFKA_CONNECT_URL="${1:-http://localhost:8083}"
KAFKA_CONNECT_DIR="kafka-connect"

echo "Loading Kafka Connect sinks from $KAFKA_CONNECT_DIR to $KAFKA_CONNECT_URL"

# Check if Kafka Connect is reachable
if ! curl -s -f "$KAFKA_CONNECT_URL" > /dev/null 2>&1; then
    echo "Error: Cannot reach Kafka Connect at $KAFKA_CONNECT_URL"
    echo "Make sure Kafka Connect is running and accessible"
    exit 1
fi

# Function to load a connector
load_connector() {
    local file_path=$1
    local file_name=$(basename "$file_path")
    
    if [ ! -f "$file_path" ]; then
        echo "Error: Connector file $file_path not found"
        return 1
    fi
    
    # Extract connector name from JSON
    connector_name=$(jq -r '.name' "$file_path" 2>/dev/null)
    
    if [ -z "$connector_name" ] || [ "$connector_name" = "null" ]; then
        echo "Error: Could not extract connector name from $file_name"
        return 1
    fi
    
    echo "Loading connector: $connector_name"
    
    # Remove connection.password field for local environment
    modified_config=$(jq 'del(.config."connection.password")' "$file_path")
    
    # Load connector configuration
    response=$(echo "$modified_config" | curl -s -w "\n%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d @- \
        "$KAFKA_CONNECT_URL/connectors")
    
    # Extract HTTP status code (last line)
    http_code=$(echo "$response" | tail -n1)
    # Extract response body (all but last line)
    response_body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" = "201" ] || [ "$http_code" = "200" ]; then
        echo "Successfully loaded connector: $connector_name"
        if [ -n "$response_body" ]; then
            echo "$response_body" | jq '.' 2>/dev/null || echo "$response_body"
        fi
        return 0
    elif [ "$http_code" = "409" ]; then
        echo "Connector $connector_name already exists, updating..."
        # Extract only the config object for PUT request
        config_only=$(echo "$modified_config" | jq '.config')
        # Update existing connector
        update_response=$(echo "$config_only" | curl -s -w "\n%{http_code}" -X PUT \
            -H "Content-Type: application/json" \
            -d @- \
            "$KAFKA_CONNECT_URL/connectors/$connector_name/config")
        
        update_http_code=$(echo "$update_response" | tail -n1)
        update_response_body=$(echo "$update_response" | head -n-1)
        
        if [ "$update_http_code" = "200" ]; then
            echo "Successfully updated connector: $connector_name"
            if [ -n "$update_response_body" ]; then
                echo "$update_response_body" | jq '.' 2>/dev/null || echo "$update_response_body"
            fi
            return 0
        else
            echo "Failed to update connector $connector_name (HTTP $update_http_code)"
            echo "Response: $update_response_body"
            return 1
        fi
    else
        echo "Failed to load connector $connector_name (HTTP $http_code)"
        echo "Response: $response_body"
        return 1
    fi
}

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed"
    echo "Install it with: brew install jq (macOS) or apt-get install jq (Linux)"
    exit 1
fi

# Check if kafka-connect directory exists
if [ ! -d "$KAFKA_CONNECT_DIR" ]; then
    echo "Error: Directory $KAFKA_CONNECT_DIR not found"
    exit 1
fi

# Load all connector JSON files
echo "=== Loading Kafka Connect sinks ==="

success_count=0
failure_count=0

for sink_file in "$KAFKA_CONNECT_DIR"/*.json; do
    if [ -f "$sink_file" ]; then
        if load_connector "$sink_file"; then
            ((success_count++))
        else
            ((failure_count++))
        fi
        echo ""
    fi
done

echo "=== Connector loading complete ==="
echo "Successfully loaded: $success_count"
echo "Failed: $failure_count"

# List all connectors
echo ""
echo "=== Current connectors in Kafka Connect ==="
curl -s "$KAFKA_CONNECT_URL/connectors" | jq '.[]' 2>/dev/null || echo "No connectors found or jq not available"

if [ $failure_count -gt 0 ]; then
    exit 1
fi

