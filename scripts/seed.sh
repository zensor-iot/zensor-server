#!/bin/bash
# Seeds a running zensor-server instance with test data via its HTTP API.
# Safe to re-run: reuses an existing "Seed Tenant" / seed devices instead of
# duplicating them, and only creates evaluation rules/tasks for devices it
# creates for the first time.
set -euo pipefail

BASE_URL="${SEED_BASE_URL:-http://127.0.0.1:3000}"

if ! command -v jq &> /dev/null; then
    echo "❌ jq is required to run the seed script (brew install jq)"
    exit 1
fi

post() {
    curl -sf -X POST "${BASE_URL}${1}" -H "Content-Type: application/json" -d "${2}"
}

find_by_name() {
    # $1 = path, $2 = name
    curl -sf "${BASE_URL}${1}?limit=100" | jq -r --arg name "$2" '.data[] | select(.name == $name) | .id' | head -n1
}

echo "🌱 seeding test data into ${BASE_URL}..."

echo "  - tenant: Seed Tenant"
TENANT_ID=$(find_by_name "/v1/tenants" "Seed Tenant")
if [ -z "$TENANT_ID" ]; then
    TENANT_ID=$(post "/v1/tenants" '{"name":"Seed Tenant","email":"seed@zensor.local","description":"Seed data for local development"}' | jq -r '.id')
    echo "    created tenant_id=${TENANT_ID}"
else
    echo "    reusing tenant_id=${TENANT_ID}"
fi

seed_device() {
    # $1 = device name, $2 = create request body
    # Prints "<is_new> <id>" on stdout; status lines go to stderr so callers
    # can safely capture just the result via command substitution.
    local name="$1" body="$2" id is_new
    id=$(find_by_name "/v1/devices" "$name")
    if [ -z "$id" ]; then
        id=$(post "/v1/devices" "$body" | jq -r '.id')
        is_new=1
        echo "    created device_id=${id}" >&2
    else
        is_new=0
        echo "    reusing device_id=${id}" >&2
    fi
    post "/v1/tenants/${TENANT_ID}/adopt" "{\"device_id\":\"${id}\"}" > /dev/null
    echo "${is_new} ${id}"
}

echo "  - device: seed-device-01"
read -r DEVICE_1_NEW DEVICE_1_ID <<< "$(seed_device "seed-device-01" '{"name":"seed-device-01","display_name":"Seed Device 01","app_eui":"0000000000000001","dev_eui":"0000000000000001","app_key":"00000000000000000000000000000001"}')"

echo "  - device: seed-device-02"
read -r DEVICE_2_NEW DEVICE_2_ID <<< "$(seed_device "seed-device-02" '{"name":"seed-device-02","display_name":"Seed Device 02","app_eui":"0000000000000002","dev_eui":"0000000000000002","app_key":"00000000000000000000000000000002"}')"

if [ "$DEVICE_1_NEW" = "1" ]; then
    echo "  - evaluation rule + task on seed-device-01"
    post "/v1/devices/${DEVICE_1_ID}/evaluation-rules" '{"description":"Alert when temperature is out of range","kind":"threshold","parameters":[{"key":"metric","value":"temperature"},{"key":"lower_threshold","value":0},{"key":"upper_threshold","value":30}]}' > /dev/null
    post "/v1/devices/${DEVICE_1_ID}/tasks" '{"commands":[{"index":1,"value":1,"priority":"NORMAL","wait_for":"0s"}]}' > /dev/null
fi

if [ "$DEVICE_2_NEW" = "1" ]; then
    echo "  - scheduled task on seed-device-02"
    post "/v1/tenants/${TENANT_ID}/devices/${DEVICE_2_ID}/scheduled-tasks" '{"commands":[{"index":1,"value":1,"priority":"NORMAL","wait_for":"0s"}],"scheduling":{"type":"cron","schedule":"0 0 * * *"},"is_active":true}' > /dev/null
fi

echo ""
echo "✅ seed data ready:"
echo "   tenant:  ${TENANT_ID}  (Seed Tenant)"
echo "   device:  ${DEVICE_1_ID}  (seed-device-01 — evaluation rule + one-off task)"
echo "   device:  ${DEVICE_2_ID}  (seed-device-02 — daily cron scheduled task)"
