# Step Definitions

This file documents the step definitions for the functional tests to help IDEs and linters understand the mapping.

## Tenant Steps

### When Steps
- `I create a new tenant with name "([^"]*)" and email "([^"]*)"` → `iCreateANewTenantWithNameAndEmail(name, email string)`
- `I get the tenant by its ID` → `iGetTheTenantByItsID()`
- `I list all tenants` → `iListAllTenants()`
- `I update the tenant with a new name "([^"]*)"` → `iUpdateTheTenantWithANewName(newName string)`
- `I deactivate the tenant` → `iDeactivateTheTenant()`
- `I activate the tenant` → `iActivateTheTenant()`
- `I soft delete the tenant` → `iSoftDeleteTheTenant()`

### Given Steps
- `a tenant exists with name "([^"]*)" and email "([^"]*)"` → `aTenantExistsWithNameAndEmail(name, email string)`
- `a deactivated tenant exists with name "([^"]*)" and email "([^"]*)"` → `aDeactivatedTenantExistsWithNameAndEmail(name, email string)`

### Then Steps
- `the response should contain the tenant with name "([^"]*)"` → `theResponseShouldContainTheTenantWithName(name string)`
- `the list should contain the tenant with name "([^"]*)"` → `theListShouldContainTheTenantWithName(name string)`
- `the tenant should be soft deleted` → `theTenantShouldBeSoftDeleted()`

## Generic Steps

### When Steps
- `wait for (.*)` → `waitForDuration(duration string)`

### Then Steps
- `the response status code should be (\d+)` → `theResponseStatusCodeShouldBe(code int)`
- `the response should contain the tenant details` → `theResponseShouldContainTheTenantDetails()`
- `the response should contain the device details` → `theResponseShouldContainTheDeviceDetails()`
- `the response should contain the task details` → `theResponseShouldContainTheTaskDetails()`
- `the response should contain the scheduled task details` → `theResponseShouldContainTheScheduledTaskDetails()`
- `the response should contain the evaluation rule details` → `theResponseShouldContainTheEvaluationRuleDetails()`

## Device Steps

### When Steps
- `I create a new device with name "([^"]*)" and display name "([^"]*)"` → `iCreateANewDeviceWithNameAndDisplayName(name, displayName string)`
- `I list all devices` → `iListAllDevices()`
- `I update the device with a new display name "([^"]*)"` → `iUpdateTheDeviceWithANewDisplayName(newDisplayName string)`

### Given Steps
- `a device exists with name "([^"]*)"` → `aDeviceExistsWithName(name string)`

### Then Steps
- `the list should contain the device with name "([^"]*)"` → `theListShouldContainTheDeviceWithName(name string)`
- `the response should contain the device with display name "([^"]*)"` → `theResponseShouldContainTheDeviceWithDisplayName(displayName string)`

## Task Steps

### When Steps
- `I create a task for the device` → `iCreateATaskForTheDevice()`

## Scheduled Task Steps

### When Steps
- `I create a scheduled task for the tenant and device with schedule "([^"]*)"` → `iCreateAScheduledTaskForTheTenantAndDeviceWithSchedule(schedule string)`
- `I list all scheduled tasks for the tenant` → `iListAllScheduledTasksForTheTenant()`
- `I update the scheduled task with a new schedule "([^"]*)"` → `iUpdateTheScheduledTaskWithANewSchedule(newSchedule string)`

### Given Steps
- `a scheduled task exists for the tenant and device with schedule "([^"]*)"` → `aScheduledTaskExistsForTheTenantAndDeviceWithSchedule(schedule string)`

### Then Steps
- `the list should contain our scheduled task` → `theListShouldContainOurScheduledTask()`
- `the response should contain the scheduled task with the new schedule` → `theResponseShouldContainTheScheduledTaskWithTheNewSchedule()`

## Evaluation Rule Steps

### When Steps
- `I create an evaluation rule for the device` → `iCreateAnEvaluationRuleForTheDevice()`
- `I list all evaluation rules for the device` → `iListAllEvaluationRulesForTheDevice()`

### Given Steps
- `an evaluation rule exists for the device` → `anEvaluationRuleExistsForTheDevice()`

### Then Steps
- `the list should contain our evaluation rule` → `theListShouldContainOurEvaluationRule()` 