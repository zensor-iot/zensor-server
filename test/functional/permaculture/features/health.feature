@health
Feature: Health Endpoint
  In order to monitor the application health and version information
  As an API user
  I want to be able to check the healthz endpoint that returns version and commit information.

  Scenario: Get healthz information
    When I call the healthz endpoint
    Then the response status code should be 200
    And the response should contain status information
    And the response should contain version information
    And the response should contain commit hash information
