@evaluation_rule
Feature: Evaluation Rule Management
  In order to manage evaluation rules in Zensor Server
  As an API user
  I want to be able to create and list evaluation rules for a device.

  Scenario: Create an evaluation rule for a device
    Given a device exists with name "rule-device-001"
    When I create an evaluation rule for the device
    And wait for 250ms
    Then the response status code should be 201
    And the response should contain the evaluation rule details

  Scenario: List evaluation rules for a device
    Given a device exists with name "rule-device-002"
    And an evaluation rule exists for the device
    When I list all evaluation rules for the device
    Then the response status code should be 200
    And the list should contain our evaluation rule
