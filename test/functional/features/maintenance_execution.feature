@maintenance_execution
Feature: Maintenance Execution Management
  In order to manage maintenance executions in Zensor Server
  As an API user
  I want to be able to list, retrieve, and mark maintenance executions as completed.

  Background:
    Given a tenant exists with name "Maintenance Execution Tenant" and email "maintenance-execution@test.com"
    And a maintenance activity exists for tenant with type "water_system" and name "Water Filter Replacement"

  Scenario: List maintenance executions for an activity
    Given there are 2 maintenance executions for the activity
    When I list all maintenance executions for the activity
    Then the response status code should be 200
    And I should receive 2 executions

  Scenario: Get a maintenance execution by ID
    Given a maintenance execution exists for the activity
    When I get the maintenance execution by its ID
    Then the response status code should be 200
    And the response should contain the maintenance execution details

  Scenario: Mark a maintenance execution as completed
    Given a maintenance execution exists for the activity
    When I mark the maintenance execution as completed by "user@test.com"
    And wait for 50ms
    And I get the maintenance execution by its ID
    Then the response status code should be 200
    And the response should contain a completed maintenance execution
    And the response should contain completed_by "user@test.com"

  Scenario: Get an overdue maintenance execution
    Given an overdue maintenance execution exists for the activity
    When I get the maintenance execution by its ID
    Then the response status code should be 200
    And the response should contain an overdue maintenance execution

