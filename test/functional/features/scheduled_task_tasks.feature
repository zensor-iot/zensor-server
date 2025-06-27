@scheduled_task_tasks
Feature: Get Tasks by Scheduled Task
  As a user
  I want to retrieve all tasks created from a scheduled task
  So that I can monitor the execution history

  Background:
    Given a tenant with id "test-tenant-1"
    And a device with id "test-device-1" belonging to tenant "test-tenant-1"
    And a scheduled task with id "test-scheduled-task-1" for device "test-device-1" with schedule "0 0 * * *"

  @wip
  Scenario: Get tasks by scheduled task with pagination
    Given there are 15 tasks created from scheduled task "test-scheduled-task-1"
    When I retrieve the first 10 tasks for scheduled task "test-scheduled-task-1"
    Then I should receive 10 tasks
    And the tasks should be sorted by creation date in descending order
    And pagination information should be included

  Scenario: Get tasks by scheduled task with custom pagination
    Given there are 15 tasks created from scheduled task "test-scheduled-task-1"
    When I retrieve page 2 with 5 tasks for scheduled task "test-scheduled-task-1"
    Then I should receive 5 tasks
    And the pagination should indicate page 2
    And pagination information should be included

  Scenario: Get tasks by scheduled task when no tasks exist
    When I retrieve tasks for scheduled task "test-scheduled-task-1"
    Then I should receive 0 tasks
    And pagination information should be included

  Scenario: Get tasks by non-existent scheduled task
    When I try to retrieve tasks for non-existent scheduled task "non-existent"
    Then the operation should fail with an error

  Scenario: Get tasks by scheduled task with invalid tenant
    When I try to retrieve tasks for scheduled task "test-scheduled-task-1" using invalid tenant "invalid-tenant"
    Then the operation should fail with an error

  Scenario: Get tasks by scheduled task with invalid device
    When I try to retrieve tasks for scheduled task "test-scheduled-task-1" using invalid device "invalid-device"
    Then the operation should fail with an error
