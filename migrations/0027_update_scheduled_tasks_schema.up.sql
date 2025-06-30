-- Drop the existing scheduled_tasks source and views
DROP MATERIALIZED VIEW IF EXISTS scheduled_tasks_final;
DROP SOURCE IF EXISTS scheduled_tasks;

-- Recreate the scheduled_tasks source with updated schema
CREATE SOURCE IF NOT EXISTS scheduled_tasks
FROM KAFKA CONNECTION kafka_connection (TOPIC 'scheduled_tasks')
KEY FORMAT TEXT
VALUE FORMAT JSON
ENVELOPE UPSERT; 