-- Add last_executed_at column to scheduled_tasks table
ALTER TABLE scheduled_tasks ADD COLUMN last_executed_at TIMESTAMP WITH TIME ZONE;

-- Update the materialized view to include the new field
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
    updated_at,
    last_executed_at
FROM scheduled_tasks;

-- Create index on the new column for better query performance
CREATE INDEX idx_scheduled_tasks_last_executed_at ON scheduled_tasks(last_executed_at); 