package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"zensor-server/test/functional/driver"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/require"
)

// PaginatedResponse represents the new paginated response format
type PaginatedResponse[T any] struct {
	Data       []T `json:"data"`
	Pagination struct {
		Total  int `json:"total"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"pagination"`
}

type FeatureContext struct {
	apiDriver        *driver.APIDriver
	response         *http.Response
	responseData     map[string]any
	responseListData []map[string]any
	tenantID         string
	deviceID         string
	scheduledTaskID  string
	evaluationRuleID string
	updatedSchedule  string
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
	ctx.Step(`^wait for (.*)$`, fc.waitForDuration)
	ctx.Then(`^the response status code should be (\d+)$`, fc.theResponseStatusCodeShouldBe)
	ctx.Then(`^the response should contain the tenant details$`, fc.theResponseShouldContainTheTenantDetails)
	ctx.Then(`^the response should contain the device details$`, fc.theResponseShouldContainTheDeviceDetails)
	ctx.Then(`^the response should contain the task details$`, fc.theResponseShouldContainTheTaskDetails)
	ctx.Then(`^the response should contain the scheduled task details$`, fc.theResponseShouldContainTheScheduledTaskDetails)
	ctx.Then(`^the response should contain the evaluation rule details$`, fc.theResponseShouldContainTheEvaluationRuleDetails)
	ctx.Then(`^the tenant should be soft deleted$`, fc.theTenantShouldBeSoftDeleted)

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
	ctx.When(`^I get the device by its ID$`, fc.iGetTheDeviceByItsID)
	ctx.Then(`^the response should contain the device with display name "([^"]*)"$`, fc.theResponseShouldContainTheDeviceWithDisplayName)

	// Task steps
	ctx.When(`^I create a task for the device$`, fc.iCreateATaskForTheDevice)
	ctx.Then(`^the response should contain command details$`, fc.theResponseShouldContainCommandDetails)

	// Scheduled Task steps
	ctx.Given(`^a scheduled task exists for the tenant and device with schedule "([^"]*)"$`, fc.aScheduledTaskExistsForTheTenantAndDeviceWithSchedule)
	ctx.When(`^I create a scheduled task for the tenant and device with schedule "([^"]*)"$`, fc.iCreateAScheduledTaskForTheTenantAndDeviceWithSchedule)
	ctx.When(`^I list all scheduled tasks for the tenant$`, fc.iListAllScheduledTasksForTheTenant)
	ctx.Then(`^the list should contain our scheduled task$`, fc.theListShouldContainOurScheduledTask)
	ctx.When(`^I update the scheduled task with a new schedule "([^"]*)"$`, fc.iUpdateTheScheduledTaskWithANewSchedule)
	ctx.When(`^I get the scheduled task by its ID$`, fc.iGetTheScheduledTaskByItsID)
	ctx.Then(`^the response should contain the scheduled task with the new schedule "([^"]*)"$`, fc.theResponseShouldContainTheScheduledTaskWithTheNewSchedule)
	ctx.When(`^I delete the scheduled task$`, fc.iDeleteTheScheduledTask)
	ctx.When(`^I try to get the scheduled task by its ID$`, fc.iTryToGetTheScheduledTaskByItsID)

	// Scheduled Task Tasks steps
	ctx.Given(`^there are (\d+) tasks created from scheduled task "([^"]*)"$`, fc.thereAreTasksCreatedFromScheduledTask)
	ctx.When(`^I retrieve the first (\d+) tasks for scheduled task "([^"]*)"$`, fc.iRetrieveTheFirstTasksForScheduledTask)
	ctx.When(`^I retrieve page (\d+) with (\d+) tasks for scheduled task "([^"]*)"$`, fc.iRetrievePageWithTasksForScheduledTask)
	ctx.When(`^I retrieve tasks for scheduled task "([^"]*)"$`, fc.iRetrieveTasksForScheduledTask)
	ctx.Then(`^I should receive (\d+) tasks$`, fc.iShouldReceiveTasks)
	ctx.Then(`^the tasks should be sorted by creation date in descending order$`, fc.theTasksShouldBeSortedByCreationDateInDescendingOrder)
	ctx.Then(`^pagination information should be included$`, fc.paginationInformationShouldBeIncluded)
	ctx.Then(`^the pagination should indicate page (\d+)$`, fc.thePaginationShouldIndicatePage)
	ctx.When(`^I try to retrieve tasks for non-existent scheduled task "([^"]*)"$`, fc.iTryToRetrieveTasksForNonExistentScheduledTask)
	ctx.When(`^I try to retrieve tasks for scheduled task "([^"]*)" using invalid tenant "([^"]*)"$`, fc.iTryToRetrieveTasksForScheduledTaskUsingInvalidTenant)
	ctx.When(`^I try to retrieve tasks for scheduled task "([^"]*)" using invalid device "([^"]*)"$`, fc.iTryToRetrieveTasksForScheduledTaskUsingInvalidDevice)
	ctx.Then(`^the operation should fail with an error$`, fc.theOperationShouldFailWithAnError)

	// Background steps for scheduled task tasks feature
	ctx.Given(`^a tenant with id "([^"]*)"$`, fc.aTenantWithId)
	ctx.Given(`^a device with id "([^"]*)" belonging to tenant "([^"]*)"$`, fc.aDeviceWithIdBelongingToTenant)
	ctx.Given(`^a scheduled task with id "([^"]*)" for device "([^"]*)" with schedule "([^"]*)"$`, fc.aScheduledTaskWithIdForDeviceWithSchedule)

	// Evaluation Rule steps
	ctx.Given(`^an evaluation rule exists for the device$`, fc.anEvaluationRuleExistsForTheDevice)
	ctx.When(`^I create an evaluation rule for the device$`, fc.iCreateAnEvaluationRuleForTheDevice)
	ctx.When(`^I list all evaluation rules for the device$`, fc.iListAllEvaluationRulesForTheDevice)
	ctx.Then(`^the list should contain our evaluation rule$`, fc.theListShouldContainOurEvaluationRule)

	// Device State Cache steps
	ctx.Given(`^the device has cached sensor data$`, fc.theDeviceHasCachedSensorData)
	ctx.When(`^I connect to the WebSocket endpoint$`, fc.iConnectToTheWebSocketEndpoint)
	ctx.Then(`^I should receive cached device states immediately$`, fc.iShouldReceiveCachedDeviceStatesImmediately)
	ctx.Then(`^the cached states should contain the device data$`, fc.theCachedStatesShouldContainTheDeviceData)

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		fc.t = godog.T(ctx)
		fc.require = require.New(fc.t)

		fc.reset()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		fc.cleanupWebSocket()
		return ctx, err
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
	fc.updatedSchedule = ""
}

func (fc *FeatureContext) decodeBody(body io.ReadCloser, target any) error {
	return json.NewDecoder(body).Decode(target)
}

func (fc *FeatureContext) decodePaginatedResponse(body *http.Response) ([]map[string]any, error) {
	var paginatedResp PaginatedResponse[map[string]any]
	if err := fc.decodeBody(body.Body, &paginatedResp); err != nil {
		return nil, fmt.Errorf("failed to decode paginated response: %w", err)
	}
	return paginatedResp.Data, nil
}
