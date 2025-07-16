Feature: Tenant Configuration Management
  As a system administrator
  I want to manage tenant configurations
  So that I can set timezone preferences for each tenant

  Background:
    Given I have a tenant with id "test-tenant-1"
    And I have a tenant with id "test-tenant-2"

  Scenario: Create tenant configuration with valid timezone
    When I create a tenant configuration for tenant "test-tenant-1" with timezone "America/New_York"
    Then the tenant configuration should be created successfully
    And the response should contain timezone "America/New_York"

  Scenario: Get tenant configuration
    Given I have a tenant configuration for tenant "test-tenant-1" with timezone "America/New_York"
    When I get the tenant configuration for tenant "test-tenant-1"
    Then the tenant configuration should be retrieved successfully
    And the response should contain timezone "America/New_York"

  Scenario: Update tenant configuration with valid timezone
    Given I have a tenant configuration for tenant "test-tenant-1" with timezone "America/New_York"
    When I update the tenant configuration for tenant "test-tenant-1" with timezone "Europe/London"
    Then the tenant configuration should be updated successfully
    And the response should contain timezone "Europe/London"

  Scenario: Create tenant configuration with invalid timezone
    When I create a tenant configuration for tenant "test-tenant-1" with timezone "Invalid/Timezone"
    Then the response should be "400 Bad Request"
    And the error message should be "invalid timezone"

  Scenario: Update tenant configuration with invalid timezone
    Given I have a tenant configuration for tenant "test-tenant-1" with timezone "America/New_York"
    When I update the tenant configuration for tenant "test-tenant-1" with timezone "Invalid/Timezone"
    Then the response should be "400 Bad Request"
    And the error message should be "invalid timezone"

  Scenario: Create duplicate tenant configuration
    Given I have a tenant configuration for tenant "test-tenant-1" with timezone "America/New_York"
    When I create a tenant configuration for tenant "test-tenant-1" with timezone "Europe/London"
    Then the response should be "500 Internal Server Error"
    And the error message should be "failed to create tenant configuration"

  Scenario: Update tenant configuration with wrong version
    Given I have a tenant configuration for tenant "test-tenant-1" with timezone "America/New_York"
    When I update the tenant configuration for tenant "test-tenant-1" with timezone "Europe/London" and version 999
    Then the response should be "500 Internal Server Error"
    And the error message should be "failed to update tenant configuration"

  Scenario: Get non-existent tenant configuration
    When I get the tenant configuration for tenant "non-existent-tenant"
    Then the response should be "404 Not Found"
    And the error message should be "tenant configuration not found"
