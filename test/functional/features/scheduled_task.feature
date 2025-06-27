@scheduled_task
Feature: Scheduled Task Management
  In order to manage scheduled tasks in Zensor Server
  As an API user
  I want to be able to create, list, and update scheduled tasks for a tenant.

  Scenario: Create a scheduled task
    Given a tenant exists with name "ScheduledTaskTenant" and email "stt@example.com"
    And a device exists with name "scheduled-task-device-001"
    When I create a scheduled task for the tenant and device with schedule "* * * * *"
    And wait for 250ms
    Then the response status code should be 201
    And the response should contain the scheduled task details

  Scenario: List scheduled tasks
    Given a tenant exists with name "ScheduledTaskTenantList" and email "sttl@example.com"
    And a device exists with name "scheduled-task-device-002"
    And a scheduled task exists for the tenant and device with schedule "* * * * *"
    When I list all scheduled tasks for the tenant
    Then the response status code should be 200
    And the list should contain our scheduled task

  @wip
  Scenario: Update a scheduled task
    Given a tenant exists with name "ScheduledTaskTenantUpdate" and email "sttu@example.com"
    And a device exists with name "scheduled-task-device-003"
    And a scheduled task exists for the tenant and device with schedule "* * * * *"
    When I update the scheduled task with a new schedule "*/5 * * * *"
    And wait for 250ms
    And I get the scheduled task by its ID
    Then the response status code should be 200
    And the response should contain the scheduled task with the new schedule
