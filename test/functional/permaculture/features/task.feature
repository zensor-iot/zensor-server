@task
Feature: Task Management
  In order to manage tasks in Zensor Server
  As an API user
  I want to be able to create tasks for devices.

  Scenario: Create a task for a device
    Given a device exists with name "task-device-001"
    And wait for 250ms
    When I create a task for the device
    And wait for 250ms
    Then the response status code should be 201
    And the response should contain the task details
    And the response should contain command details

