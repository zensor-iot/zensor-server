Feature: Get Tasks by Scheduled Task
  As a user
  I want to retrieve all tasks created from a scheduled task
  So that I can monitor the execution history

  Background:
    Given a tenant with id "test-tenant-1"
    And a device with id "test-device-1" belonging to tenant "test-tenant-1"
    And a scheduled task with id "test-scheduled-task-1" for device "test-device-1" with schedule "0 0 * * *"

  Scenario: Get tasks by scheduled task with pagination
    Given there are 15 tasks created from scheduled task "test-scheduled-task-1"
    When I request "GET /v1/tenants/test-tenant-1/devices/test-device-1/scheduled-tasks/test-scheduled-task-1/tasks?page=1&limit=10"
    Then the response status should be 200
    And the response should contain pagination metadata
    And the response should contain 10 tasks
    And the tasks should be sorted by created_at in descending order

  Scenario: Get tasks by scheduled task with custom pagination
    Given there are 15 tasks created from scheduled task "test-scheduled-task-1"
    When I request "GET /v1/tenants/test-tenant-1/devices/test-device-1/scheduled-tasks/test-scheduled-task-1/tasks?page=2&limit=5"
    Then the response status should be 200
    And the response should contain pagination metadata
    And the response should contain 5 tasks
    And the pagination should show page 2

  Scenario: Get tasks by scheduled task when no tasks exist
    When I request "GET /v1/tenants/test-tenant-1/devices/test-device-1/scheduled-tasks/test-scheduled-task-1/tasks"
    Then the response status should be 200
    And the response should contain pagination metadata
    And the response should contain 0 tasks

  Scenario: Get tasks by non-existent scheduled task
    When I request "GET /v1/tenants/test-tenant-1/devices/test-device-1/scheduled-tasks/non-existent/tasks"
    Then the response status should be 500

  Scenario: Get tasks by scheduled task with invalid tenant
    When I request "GET /v1/tenants/invalid-tenant/devices/test-device-1/scheduled-tasks/test-scheduled-task-1/tasks"
    Then the response status should be 500

  Scenario: Get tasks by scheduled task with invalid device
    When I request "GET /v1/tenants/test-tenant-1/devices/invalid-device/scheduled-tasks/test-scheduled-task-1/tasks"
    Then the response status should be 500
