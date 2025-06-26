@device
Feature: Device Management
  In order to manage devices in Zensor Server
  As an API user
  I want to be able to create, list, and update devices.

  Scenario: Create a new device
    When I create a new device with name "sensor-001" and display name "Temperature Sensor 1"
    And wait for 250ms
    Then the response status code should be 201
    And the response should contain the device details

  Scenario: List devices
    Given a device exists with name "sensor-list"
    When I list all devices
    Then the response status code should be 200
    And the list should contain the device with name "sensor-list"

  Scenario: Update a device's display name
    Given a device exists with name "sensor-002"
    When I update the device with a new display name "Humidity Sensor"
    And wait for 250ms
    And I get the device by its ID
    Then the response status code should be 200
    And the response should contain the device with display name "Humidity Sensor"
