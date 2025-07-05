#!/bin/bash

# Script to update Avro schemas in the schema registry
# This script deletes existing schemas and registers new ones with logical timestamp types

SCHEMA_REGISTRY_URL="http://localhost:8081"

echo "Updating Avro schemas in schema registry..."

# Function to delete all versions of a subject
delete_subject() {
    local subject=$1
    echo "Deleting subject: $subject"
    curl -X DELETE "$SCHEMA_REGISTRY_URL/subjects/$subject" -s | jq .
}

# Function to register a new schema
register_schema() {
    local subject=$1
    local schema_file=$2
    echo "Registering schema for subject: $subject"
    curl -X POST "$SCHEMA_REGISTRY_URL/subjects/$subject/versions" \
        -H "Content-Type: application/vnd.schemaregistry.v1+json" \
        -d @"$schema_file" -s | jq .
}

# Delete existing subjects
echo "Deleting existing subjects..."
delete_subject "commands-value"
delete_subject "tasks-value"
delete_subject "devices-value"
delete_subject "scheduled-tasks-value"
delete_subject "tenants-value"
delete_subject "evaluation-rules-value"

echo "Waiting for deletion to complete..."
sleep 2

# Create temporary schema files
TEMP_DIR=$(mktemp -d)

# Command schema
cat > "$TEMP_DIR/command_schema.json" << 'EOF'
{
  "schema": "{\"type\":\"record\",\"name\":\"Command\",\"fields\":[{\"name\":\"id\",\"type\":\"string\"},{\"name\":\"version\",\"type\":\"int\"},{\"name\":\"device_name\",\"type\":\"string\"},{\"name\":\"device_id\",\"type\":\"string\"},{\"name\":\"task_id\",\"type\":\"string\"},{\"name\":\"payload\",\"type\":{\"type\":\"record\",\"name\":\"CommandPayload\",\"fields\":[{\"name\":\"index\",\"type\":\"int\"},{\"name\":\"value\",\"type\":\"int\"}]}},{\"name\":\"dispatch_after\",\"type\":{\"type\":\"long\",\"logicalType\":\"timestamp-millis\"}},{\"name\":\"port\",\"type\":\"int\"},{\"name\":\"priority\",\"type\":\"string\"},{\"name\":\"created_at\",\"type\":{\"type\":\"long\",\"logicalType\":\"timestamp-millis\"}},{\"name\":\"ready\",\"type\":\"boolean\"},{\"name\":\"sent\",\"type\":\"boolean\"},{\"name\":\"sent_at\",\"type\":{\"type\":\"long\",\"logicalType\":\"timestamp-millis\"}}]}"
}
EOF

# Task schema
cat > "$TEMP_DIR/task_schema.json" << 'EOF'
{
  "schema": "{\"type\":\"record\",\"name\":\"Task\",\"fields\":[{\"name\":\"id\",\"type\":\"string\"},{\"name\":\"device_id\",\"type\":\"string\"},{\"name\":\"scheduled_task_id\",\"type\":[\"null\",\"string\"]},{\"name\":\"version\",\"type\":\"long\"},{\"name\":\"created_at\",\"type\":{\"type\":\"long\",\"logicalType\":\"timestamp-millis\"}},{\"name\":\"updated_at\",\"type\":{\"type\":\"long\",\"logicalType\":\"timestamp-millis\"}}]}"
}
EOF

# Device schema
cat > "$TEMP_DIR/device_schema.json" << 'EOF'
{
  "schema": "{\"type\":\"record\",\"name\":\"Device\",\"fields\":[{\"name\":\"id\",\"type\":\"string\"},{\"name\":\"version\",\"type\":\"int\"},{\"name\":\"name\",\"type\":\"string\"},{\"name\":\"display_name\",\"type\":\"string\"},{\"name\":\"app_eui\",\"type\":\"string\"},{\"name\":\"dev_eui\",\"type\":\"string\"},{\"name\":\"app_key\",\"type\":\"string\"},{\"name\":\"tenant_id\",\"type\":[\"null\",\"string\"]},{\"name\":\"last_message_received_at\",\"type\":[\"null\",{\"type\":\"long\",\"logicalType\":\"timestamp-millis\"}]},{\"name\":\"created_at\",\"type\":{\"type\":\"long\",\"logicalType\":\"timestamp-millis\"}},{\"name\":\"updated_at\",\"type\":{\"type\":\"long\",\"logicalType\":\"timestamp-millis\"}}]}"
}
EOF

# ScheduledTask schema
cat > "$TEMP_DIR/scheduled_task_schema.json" << 'EOF'
{
  "schema": "{\"type\":\"record\",\"name\":\"ScheduledTask\",\"fields\":[{\"name\":\"id\",\"type\":\"string\"},{\"name\":\"version\",\"type\":\"long\"},{\"name\":\"tenant_id\",\"type\":\"string\"},{\"name\":\"device_id\",\"type\":\"string\"},{\"name\":\"command_templates\",\"type\":\"string\"},{\"name\":\"schedule\",\"type\":\"string\"},{\"name\":\"is_active\",\"type\":\"boolean\"},{\"name\":\"created_at\",\"type\":{\"type\":\"long\",\"logicalType\":\"timestamp-millis\"}},{\"name\":\"updated_at\",\"type\":{\"type\":\"long\",\"logicalType\":\"timestamp-millis\"}},{\"name\":\"last_executed_at\",\"type\":[\"null\",{\"type\":\"long\",\"logicalType\":\"timestamp-millis\"}]},{\"name\":\"deleted_at\",\"type\":[\"null\",\"string\"]}]}"
}
EOF

# Tenant schema
cat > "$TEMP_DIR/tenant_schema.json" << 'EOF'
{
  "schema": "{\"type\":\"record\",\"name\":\"Tenant\",\"fields\":[{\"name\":\"id\",\"type\":\"string\"},{\"name\":\"version\",\"type\":\"int\"},{\"name\":\"name\",\"type\":\"string\"},{\"name\":\"email\",\"type\":\"string\"},{\"name\":\"description\",\"type\":\"string\"},{\"name\":\"is_active\",\"type\":\"boolean\"},{\"name\":\"created_at\",\"type\":{\"type\":\"long\",\"logicalType\":\"timestamp-millis\"}},{\"name\":\"updated_at\",\"type\":{\"type\":\"long\",\"logicalType\":\"timestamp-millis\"}},{\"name\":\"deleted_at\",\"type\":[\"null\",\"string\"]}]}"
}
EOF

# EvaluationRule schema
cat > "$TEMP_DIR/evaluation_rule_schema.json" << 'EOF'
{
  "schema": "{\"type\":\"record\",\"name\":\"EvaluationRule\",\"fields\":[{\"name\":\"id\",\"type\":\"string\"},{\"name\":\"device_id\",\"type\":\"string\"},{\"name\":\"version\",\"type\":\"int\"},{\"name\":\"description\",\"type\":\"string\"},{\"name\":\"kind\",\"type\":\"string\"},{\"name\":\"enabled\",\"type\":\"boolean\"},{\"name\":\"parameters\",\"type\":\"string\"},{\"name\":\"created_at\",\"type\":{\"type\":\"long\",\"logicalType\":\"timestamp-millis\"}},{\"name\":\"updated_at\",\"type\":{\"type\":\"long\",\"logicalType\":\"timestamp-millis\"}}]}"
}
EOF

echo "Registering new schemas..."

register_schema "commands-value" "$TEMP_DIR/command_schema.json"
register_schema "tasks-value" "$TEMP_DIR/task_schema.json"
register_schema "devices-value" "$TEMP_DIR/device_schema.json"
register_schema "scheduled-tasks-value" "$TEMP_DIR/scheduled_task_schema.json"
register_schema "tenants-value" "$TEMP_DIR/tenant_schema.json"
register_schema "evaluation-rules-value" "$TEMP_DIR/evaluation_rule_schema.json"

# Clean up temporary files
rm -rf "$TEMP_DIR"

echo "Schema update complete!"
echo ""
echo "Current subjects:"
curl -s "$SCHEMA_REGISTRY_URL/subjects" | jq . 