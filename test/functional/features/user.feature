@user
Feature: User-Tenant Association
  In order to manage user-tenant associations in Zensor Server
  As an API user
  I want to be able to associate users with tenants and retrieve their associations.

  Scenario: Associate a user with tenants
    Given a tenant exists with name "ACME Corp" and email "contact@acme.com"
    And another tenant exists with name "ACME Corp 2" and email "contact2@acme.com"
    When I associate user "user-123" with tenants
    And wait for 50ms
    Then the response status code should be 200

  Scenario: Retrieve user-tenant associations
    Given a tenant exists with name "ACME Corp" and email "contact@acme.com"
    And another tenant exists with name "ACME Corp 2" and email "contact2@acme.com"
    And user "user-456" is associated with tenants
    When I get the user "user-456"
    Then the response status code should be 200
    And the response should contain the user with id "user-456"
    And the response should contain exactly 2 tenants

  Scenario: Update user-tenant associations
    Given a tenant exists with name "ACME Corp" and email "contact@acme.com"
    And another tenant exists with name "ACME Corp 2" and email "contact2@acme.com"
    And a third tenant exists with name "ACME Corp 3" and email "contact3@acme.com"
    And user "user-789" is associated with 2 tenants
    When I update user "user-789" with different tenants
    And wait for 50ms
    Then the response status code should be 200
    And I get the user "user-789"
    Then the response should contain the user with id "user-789"
    And the response should contain exactly 3 tenants

  Scenario: Associate user with empty tenant list
    When I associate user "user-empty" with empty tenant list
    And wait for 50ms
    Then the response status code should be 200

  Scenario: Attempt to associate user with non-existent tenant
    When I attempt to associate user "user-invalid" with non-existent tenant
    Then the response status code should be 400

  Scenario: Retrieve user that does not exist
    When I get the user "non-existent-user"
    Then the response status code should be 404

  Scenario: Associate user with valid and invalid tenants mixed
    Given a tenant exists with name "ACME Corp" and email "contact@acme.com"
    When I attempt to associate user "user-mixed" with mixed tenant list
    Then the response status code should be 400
