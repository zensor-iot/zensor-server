@scheduled_task_interval
Feature: Interval-Based Scheduled Task Management
  In order to manage scheduled tasks with interval-based scheduling
  As an API user
  I want to be able to create scheduled tasks that run every X days at a specific time.

  Scenario: Create a scheduled task with interval scheduling
    Given a tenant exists with name "IntervalTenant" and email "interval@example.com"
    And a device exists with name "interval-device-001"
    When I create a scheduled task with:
      | parameter        | value      |
      | command_index    |          1 |
      | command_value    |        100 |
      | command_priority | NORMAL     |
      | command_wait_for |         0s |
      | scheduling_type  | interval   |
      | initial_day      | 2024-01-15 |
      | day_interval     |          2 |
      | execution_time   |      02:00 |
      | is_active        | true       |
    Then the response status code should be 201
    And the response should contain the scheduled task details with interval scheduling

  Scenario: Create a scheduled task with daily interval
    Given a tenant exists with name "DailyTenant" and email "daily@example.com"
    And a device exists with name "daily-device-001"
    When I create a scheduled task with:
      | parameter        | value      |
      | command_index    |          1 |
      | command_value    |        200 |
      | command_priority | HIGH       |
      | command_wait_for |         5s |
      | scheduling_type  | interval   |
      | initial_day      | 2024-01-01 |
      | day_interval     |          1 |
      | execution_time   |      01:00 |
      | is_active        | true       |
    Then the response status code should be 201
    And the response should contain the scheduled task with next execution time

  Scenario: Create a scheduled task with 3-day interval
    Given a tenant exists with name "ThreeDayTenant" and email "threeday@example.com"
    And a device exists with name "threeday-device-001"
    When I create a scheduled task with:
      | parameter        | value      |
      | command_index    |          1 |
      | command_value    |        300 |
      | command_priority | NORMAL     |
      | command_wait_for |         0s |
      | scheduling_type  | interval   |
      | initial_day      | 2024-01-10 |
      | day_interval     |          3 |
      | execution_time   |      15:00 |
      | is_active        | true       |
    Then the response status code should be 201
    And the response should contain the scheduled task details with 3-day interval

  Scenario: Create a scheduled task with invalid interval configuration
    Given a tenant exists with name "InvalidTenant" and email "invalid@example.com"
    And a device exists with name "invalid-device-001"
    When I create a scheduled task with:
      | parameter        | value    |
      | command_index    |        1 |
      | command_value    |      100 |
      | command_priority | NORMAL   |
      | command_wait_for |       0s |
      | scheduling_type  | interval |
      | day_interval     |        2 |
      | execution_time   |    02:00 |
      | is_active        | true     |
    Then the response status code should be 400
    And the response should contain an error about missing initial_day

  Scenario: Update scheduled task from cron to interval scheduling
    Given a tenant exists with name "UpdateTenant" and email "update@example.com"
    And a device exists with name "update-device-001"
    And a scheduled task exists for the tenant and device with schedule "* * * * *"
    When I update the scheduled task with:
      | parameter       | value      |
      | scheduling_type | interval   |
      | initial_day     | 2024-02-01 |
      | day_interval    |          5 |
      | execution_time  |      10:30 |
    Then the response status code should be 200
    And the response should contain the updated scheduled task with interval scheduling
