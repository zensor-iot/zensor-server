@tenant
Feature: Tenant Management
  In order to manage tenants in Zensor Server
  As an API user
  I want to be able to create, retrieve, update, and delete tenants.

  Scenario: Create a new tenant
    When I create a new tenant with name "ACME Corp" and email "contact@acme.com"
    And wait for 50ms
    Then the response status code should be 201
    And the response should contain the tenant details

  Scenario: List all tenants
    Given a tenant exists with name "ACME Corp List" and email "contact@acmelist.com"
    When I list all tenants
    Then the response status code should be 200
    And the list should contain the tenant with name "ACME Corp List"

  Scenario: Update a tenant
    Given a tenant exists with name "ACME Corp Update" and email "contact@acmeupdate.com"
    When I update the tenant with a new name "ACME Inc."
    And wait for 50ms
    And I get the tenant by its ID
    Then the response status code should be 200
    And the response should contain the tenant with name "ACME Inc."

  Scenario: Deactivate a tenant
    Given a tenant exists with name "ACME Corp Deactivate" and email "contact@acmedeactivate.com"
    When I deactivate the tenant
    And wait for 50ms
    And I get the tenant by its ID
    Then the response status code should be 200
    And the response should contain the tenant details

  @wip
  Scenario: Activate a tenant
    Given a deactivated tenant exists with name "ACME Corp Activate" and email "contact@acmeactivate.com"
    When I activate the tenant
    And wait for 50ms
    And I get the tenant by its ID
    Then the response status code should be 200
    And the response should contain the tenant details

  Scenario: Soft delete a tenant
    Given a tenant exists with name "ACME Corp SoftDelete" and email "contact@acmesoftdelete.com"
    When I soft delete the tenant
    And wait for 50ms
    And I get the tenant by its ID
    Then the response status code should be 200
    And the tenant should be soft deleted
