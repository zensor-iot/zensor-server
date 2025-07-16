#!/bin/bash

# Script to register Avro schemas from files to the schema registry
# Usage: ./scripts/register_schemas_from_files.sh [schema-registry-url] [--token <token>]

SCHEMA_REGISTRY_URL="http://localhost:8081"
TOKEN=""
SCHEMAS_DIR="schemas"

# Parse arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --token)
      TOKEN="$2"
      shift 2
      ;;
    *)
      SCHEMA_REGISTRY_URL="$1"
      shift
      ;;
  esac
done

HEADER_TOKEN=""
if [[ -n "$TOKEN" ]]; then
  HEADER_TOKEN="-H \"X-Auth-Token: $TOKEN\""
fi

echo "Registering schemas from $SCHEMAS_DIR to $SCHEMA_REGISTRY_URL"

# Function to register a schema
register_schema() {
    local schema_name=$1
    local file_name=$2
    local subject="${schema_name}-value"
    
    echo "Registering schema: $subject"
    
    # Read schema from file
    if [ ! -f "$SCHEMAS_DIR/$file_name" ]; then
        echo "Error: Schema file $SCHEMAS_DIR/$file_name not found"
        return 1
    fi
    
    # Prepare payload
    payload=$(jq -n --arg schema "$(jq -c . < $SCHEMAS_DIR/$file_name)" '{"schema": $schema}')

    # Debug output for the first schema only
    if [ "$subject" = "device_commands-value" ]; then
        echo "--- DEBUG: Payload for $subject ---"
        echo "$payload"
        echo "--- DEBUG: Curl command ---"
        echo "curl -s -X POST -H 'Content-Type: application/vnd.schemaregistry.v1+json' $HEADER_TOKEN -d @payload.json '$SCHEMA_REGISTRY_URL/subjects/$subject/versions'"
    fi

    # Register schema with proper JSON formatting
    response=$(echo "$payload" | \
        eval curl -s -X POST \
        -H "Content-Type: application/vnd.schemaregistry.v1+json" \
        $HEADER_TOKEN \
        -d @- \
        "$SCHEMA_REGISTRY_URL/subjects/$subject/versions")
    
    # Check if response contains an error
    if echo "$response" | grep -q "error_code\|Unauthorized\|Forbidden"; then
        echo "Failed to register $subject"
        echo "Response: $response"
        return 1
    elif echo "$response" | grep -q "id"; then
        echo "Successfully registered $subject"
        echo "Response: $response"
    else
        echo "Unexpected response for $subject"
        echo "Response: $response"
        return 1
    fi
}

# Register all schemas
echo "=== Registering schemas ==="

register_schema "device_commands" "device_command.avsc"
register_schema "tasks" "task.avsc"
register_schema "devices" "device.avsc"
register_schema "scheduled_tasks" "scheduled_task.avsc"
register_schema "tenants" "tenant.avsc"
register_schema "evaluation_rules" "evaluation_rule.avsc"
register_schema "tenant_configurations" "tenant_configuration.avsc"

echo "=== Schema registration complete ==="

# List all subjects
echo "=== Current subjects in registry ==="
eval curl -s $HEADER_TOKEN "$SCHEMA_REGISTRY_URL/subjects" | jq '.[]' 2>/dev/null || echo "No subjects found or jq not available" 