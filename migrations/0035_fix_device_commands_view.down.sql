-- Revert device_commands_final view to simple structure
DROP MATERIALIZED VIEW IF EXISTS device_commands_final;
CREATE MATERIALIZED VIEW device_commands_final AS
SELECT *
FROM device_commands;