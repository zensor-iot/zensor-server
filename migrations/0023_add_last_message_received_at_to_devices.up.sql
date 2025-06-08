DROP MATERIALIZED VIEW IF EXISTS devices_final;

CREATE MATERIALIZED VIEW devices_final AS
  SELECT
    (data->>'id')::uuid AS id,
    (data->>'version')::integer AS version,
    data->>'name' AS name,
    data->>'display_name' AS display_name,
    data->>'app_eui' AS app_eui,
    data->>'dev_eui' AS dev_eui,
    data->>'app_key' AS app_key,
    (data->>'tenant_id')::uuid AS tenant_id,
    (data->>'last_message_received_at')::timestamp AS last_message_received_at,
    (data->>'created_at')::timestamp AS created_at,
    (data->>'updated_at')::timestamp AS updated_at
  FROM devices; 