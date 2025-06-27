-- Drop the existing tasks views
DROP MATERIALIZED VIEW IF EXISTS tasks_final;
DROP MATERIALIZED VIEW IF EXISTS tasks;

-- Recreate the tasks view with the new scheduled_task_id field
CREATE MATERIALIZED VIEW IF NOT EXISTS tasks AS
  SELECT *
  FROM tasks; 