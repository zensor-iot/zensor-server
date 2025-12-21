@maintenance_activity
Feature: Maintenance Activity Management
  In order to manage maintenance activities in Zensor Server
  As an API user
  I want to be able to create, retrieve, update, delete, activate, and deactivate maintenance activities.

  Scenario: Create a new maintenance activity with predefined type
    Given a tenant exists with name "Maintenance Tenant" and email "maintenance@test.com"
    When I create a maintenance activity for tenant with type "water_system" and name "Water Filter Replacement"
    And wait for 50ms
    Then the response status code should be 201
    And the response should contain the maintenance activity details

  Scenario: Create a new maintenance activity with custom type
    Given a tenant exists with name "Maintenance Tenant Custom" and email "maintenance-custom@test.com"
    When I create a maintenance activity for tenant with custom type "custom_filter" and name "Custom Filter"
    And wait for 50ms
    Then the response status code should be 201
    And the response should contain the maintenance activity details

  Scenario: List all maintenance activities for a tenant
    Given a tenant exists with name "Maintenance Tenant List" and email "maintenance-list@test.com"
    And a maintenance activity exists for tenant with type "car" and name "Oil Change"
    When I list all maintenance activities for the tenant
    Then the response status code should be 200
    And the list should contain the maintenance activity with name "Oil Change"

  Scenario: Get a maintenance activity by ID
    Given a tenant exists with name "Maintenance Tenant Get" and email "maintenance-get@test.com"
    And a maintenance activity exists for tenant with type "pets" and name "Vet Visit"
    When I get the maintenance activity by its ID
    Then the response status code should be 200
    And the response should contain the maintenance activity with name "Vet Visit"

  Scenario: Update a maintenance activity
    Given a tenant exists with name "Maintenance Tenant Update" and email "maintenance-update@test.com"
    And a maintenance activity exists for tenant with type "water_system" and name "Initial Name"
    When I update the maintenance activity with name "Updated Name"
    And wait for 50ms
    And I get the maintenance activity by its ID
    Then the response status code should be 200
    And the response should contain the maintenance activity with name "Updated Name"

  Scenario: Activate a maintenance activity
    Given a tenant exists with name "Maintenance Tenant Activate" and email "maintenance-activate@test.com"
    And a deactivated maintenance activity exists for tenant with type "car" and name "Car Service"
    When I activate the maintenance activity
    And wait for 50ms
    And I get the maintenance activity by its ID
    Then the response status code should be 200
    And the response should contain an active maintenance activity

  Scenario: Deactivate a maintenance activity
    Given a tenant exists with name "Maintenance Tenant Deactivate" and email "maintenance-deactivate@test.com"
    And a maintenance activity exists for tenant with type "water_system" and name "Water Service"
    When I deactivate the maintenance activity
    And wait for 50ms
    And I get the maintenance activity by its ID
    Then the response status code should be 200
    And the response should contain an inactive maintenance activity

  Scenario: Delete a maintenance activity
    Given a tenant exists with name "Maintenance Tenant Delete" and email "maintenance-delete@test.com"
    And a maintenance activity exists for tenant with type "pets" and name "Pet Grooming"
    When I delete the maintenance activity
    And wait for 50ms
    Then the response status code should be 204
