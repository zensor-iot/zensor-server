@web_push
Feature: Web Push configuration
  In order to subscribe browsers to push notifications
  As a web client
  I want to retrieve the VAPID public key from the server.

  Scenario: Retrieve the VAPID public key
    When I request the VAPID public key
    Then the response status code should be 200
    And the response should contain a VAPID public key
