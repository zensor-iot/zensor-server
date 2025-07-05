#!/bin/bash

# Script to register Avro schemas from files to the schema registry
# Usage: ./scripts/register_schemas_from_files.sh [schema-registry-url]

SCHEMA_REGISTRY_URL=${1:-"http://localhost:8081"}
SCHEMAS_DIR="schemas"

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
    
    # Register schema
    response=$(curl -s -X POST \
        -H "Content-Type: application/vnd.schemaregistry.v1+json" \
        -d @$SCHEMAS_DIR/$file_name \
        "$SCHEMA_REGISTRY_URL/subjects/$subject/versions")
    
    if [ $? -eq 0 ]; then
        echo "Successfully registered $subject"
        echo "Response: $response"
    else
        echo "Failed to register $subject"
        return 1
    fi
}

# Register all schemas
echo "=== Registering schemas ==="

register_schema "commands" "command.avsc"
register_schema "tasks" "task.avsc"
register_schema "devices" "device.avsc"
register_schema "scheduled-tasks" "scheduled_task.avsc"
register_schema "tenants" "tenant.avsc"
register_schema "evaluation-rules" "evaluation_rule.avsc"

echo "=== Schema registration complete ==="

# List all subjects
echo "=== Current subjects in registry ==="
curl -s "$SCHEMA_REGISTRY_URL/subjects" | jq '.[]' 2>/dev/null || echo "No subjects found or jq not available" 