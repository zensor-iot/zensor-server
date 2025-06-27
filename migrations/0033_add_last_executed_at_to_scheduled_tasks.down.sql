-- Drop the index on last_executed_at
DROP INDEX IF EXISTS idx_scheduled_tasks_last_executed_at;

-- Update the materialized view to exclude the field
DROP MATERIALIZED VIEW IF EXISTS scheduled_tasks_view_structured;
CREATE MATERIALIZED VIEW scheduled_tasks_view_structured AS
SELECT 
    id,
    version,
    tenant_id,
    device_id,
    command_templates,
    schedule,
    is_active,
    created_at,
    updated_at
FROM scheduled_tasks;

-- Remove last_executed_at column from scheduled_tasks table
ALTER TABLE scheduled_tasks DROP COLUMN IF EXISTS last_executed_at; 