package steps

import (
	"context"
	"fmt"
	"net/http"
	"zensor-server/test/functional/driver"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/require"
)

type FeatureContext struct {
	apiDriver        *driver.APIDriver
	response         *http.Response
	responseData     map[string]any
	responseListData []map[string]any
	tenantID         string
	deviceID         string
	scheduledTaskID  string
	evaluationRuleID string
	require          *require.Assertions
	t                godog.TestingT
}

func NewFeatureContext() *FeatureContext {
	return &FeatureContext{
		apiDriver: driver.NewAPIDriver("http://localhost:3000"),
	}
}

func (fc *FeatureContext) RegisterSteps(ctx *godog.ScenarioContext) {
	// Generic steps
	ctx.Given(`^the service is running$`, fc.theServiceIsRunning)
	ctx.Then(`^the response status code should be (\d+)$`, fc.theResponseStatusCodeShouldBe)
	ctx.Then(`^the response should contain the tenant details$`, fc.theResponseShouldContainTheTenantDetails)
	ctx.Then(`^the response should contain the device details$`, fc.theResponseShouldContainTheDeviceDetails)
	ctx.Then(`^the response should contain the task details$`, fc.theResponseShouldContainTheTaskDetails)
	ctx.Then(`^the response should contain the scheduled task details$`, fc.theResponseShouldContainTheScheduledTaskDetails)
	ctx.Then(`^the response should contain the evaluation rule details$`, fc.theResponseShouldContainTheEvaluationRuleDetails)

	// Tenant steps
	ctx.When(`^I create a new tenant with name "([^"]*)" and email "([^"]*)"$`, fc.iCreateANewTenantWithNameAndEmail)
	ctx.Given(`^a tenant exists with name "([^"]*)" and email "([^"]*)"$`, fc.aTenantExistsWithNameAndEmail)
	ctx.Given(`^a deactivated tenant exists with name "([^"]*)" and email "([^"]*)"$`, fc.aDeactivatedTenantExistsWithNameAndEmail)
	ctx.When(`^I get the tenant by its ID$`, fc.iGetTheTenantByItsID)
	ctx.Then(`^the response should contain the tenant with name "([^"]*)"$`, fc.theResponseShouldContainTheTenantWithName)
	ctx.When(`^I list all tenants$`, fc.iListAllTenants)
	ctx.Then(`^the list should contain the tenant with name "([^"]*)"$`, fc.theListShouldContainTheTenantWithName)
	ctx.When(`^I update the tenant with a new name "([^"]*)"$`, fc.iUpdateTheTenantWithANewName)
	ctx.When(`^I deactivate the tenant$`, fc.iDeactivateTheTenant)
	ctx.When(`^I activate the tenant$`, fc.iActivateTheTenant)
	ctx.When(`^I soft delete the tenant$`, fc.iSoftDeleteTheTenant)

	// Device steps
	ctx.When(`^I create a new device with name "([^"]*)" and display name "([^"]*)"$`, fc.iCreateANewDeviceWithNameAndDisplayName)
	ctx.Given(`^a device exists with name "([^"]*)"$`, fc.aDeviceExistsWithName)
	ctx.When(`^I list all devices$`, fc.iListAllDevices)
	ctx.Then(`^the list should contain the device with name "([^"]*)"$`, fc.theListShouldContainTheDeviceWithName)
	ctx.When(`^I update the device with a new display name "([^"]*)"$`, fc.iUpdateTheDeviceWithANewDisplayName)
	ctx.Then(`^the response should contain the device with display name "([^"]*)"$`, fc.theResponseShouldContainTheDeviceWithDisplayName)

	// Task steps
	ctx.When(`^I create a task for the device$`, fc.iCreateATaskForTheDevice)

	// Scheduled Task steps
	ctx.Given(`^a scheduled task exists for the tenant and device with schedule "([^"]*)"$`, fc.aScheduledTaskExistsForTheTenantAndDeviceWithSchedule)
	ctx.When(`^I create a scheduled task for the tenant and device with schedule "([^"]*)"$`, fc.iCreateAScheduledTaskForTheTenantAndDeviceWithSchedule)
	ctx.When(`^I list all scheduled tasks for the tenant$`, fc.iListAllScheduledTasksForTheTenant)
	ctx.Then(`^the list should contain our scheduled task$`, fc.theListShouldContainOurScheduledTask)
	ctx.When(`^I update the scheduled task with a new schedule "([^"]*)"$`, fc.iUpdateTheScheduledTaskWithANewSchedule)
	ctx.Then(`^the response should contain the scheduled task with the new schedule$`, fc.theResponseShouldContainTheScheduledTaskWithTheNewSchedule)

	// Evaluation Rule steps
	ctx.Given(`^an evaluation rule exists for the device$`, fc.anEvaluationRuleExistsForTheDevice)
	ctx.When(`^I create an evaluation rule for the device$`, fc.iCreateAnEvaluationRuleForTheDevice)
	ctx.When(`^I list all evaluation rules for the device$`, fc.iListAllEvaluationRulesForTheDevice)
	ctx.Then(`^the list should contain our evaluation rule$`, fc.theListShouldContainOurEvaluationRule)

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		t := godog.T(ctx)
		if t == nil {
			return ctx, fmt.Errorf("godog.T(ctx) returned nil")
		}
		fc.t = t
		fc.require = require.New(t)

		fc.reset()
		return ctx, nil
	})
}

func (fc *FeatureContext) reset() {
	fc.response = nil
	fc.responseData = nil
	fc.responseListData = nil
	fc.tenantID = ""
	fc.deviceID = ""
	fc.scheduledTaskID = ""
	fc.evaluationRuleID = ""
}
