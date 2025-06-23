@tenant
Feature: Tenant Management
  In order to manage tenants in Zensor Server
  As an API user
  I want to be able to create, retrieve, update, and delete tenants.

  @wip
  Scenario: Create a new tenant
    Given the service is running
    When I create a new tenant with name "ACME Corp" and email "contact@acme.com"
    Then the response status code should be 201
    And the response should contain the tenant details

  Scenario: Retrieve an existing tenant
    Given the service is running
    And a tenant exists with name "ACME Corp Get" and email "contact@acmeget.com"
    When I get the tenant by its ID
    Then the response status code should be 200
    And the response should contain the tenant with name "ACME Corp Get"

  Scenario: List all tenants
    Given the service is running
    And a tenant exists with name "ACME Corp List" and email "contact@acmelist.com"
    When I list all tenants
    Then the response status code should be 200
    And the list should contain the tenant with name "ACME Corp List"

  Scenario: Update a tenant
    Given the service is running
    And a tenant exists with name "ACME Corp Update" and email "contact@acmeupdate.com"
    When I update the tenant with a new name "ACME Inc."
    Then the response status code should be 200
    And the response should contain the tenant with name "ACME Inc."

  Scenario: Deactivate a tenant
    Given the service is running
    And a tenant exists with name "ACME Corp Deactivate" and email "contact@acmedeactivate.com"
    When I deactivate the tenant
    Then the response status code should be 204

  Scenario: Activate a tenant
    Given the service is running
    And a deactivated tenant exists with name "ACME Corp Activate" and email "contact@acmeactivate.com"
    When I activate the tenant
    Then the response status code should be 204

  Scenario: Soft delete a tenant
    Given the service is running
    And a tenant exists with name "ACME Corp SoftDelete" and email "contact@acmesoftdelete.com"
    When I soft delete the tenant
    Then the response status code should be 204
