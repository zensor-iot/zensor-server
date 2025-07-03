@device_state_cache
Feature: Device State Cache
  In order to provide quick access to device states
  As a WebSocket client
  I want to receive cached device states when I connect

  Scenario: Receive cached device states on WebSocket connection
    Given a device exists with name "cache-test-device"
    And the device has cached sensor data
    When I connect to the WebSocket endpoint
    Then I should receive cached device states immediately
    And the cached states should contain the device data
