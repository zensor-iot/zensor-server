package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"syscall"
	"time"
	"zensor-server/test/functional/permaculture/driver"

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
	apiDriver            *driver.APIDriver
	response             *http.Response
	responseData         map[string]any
	responseListData     []map[string]any
	tenantID             string
	tenantIDs            []string
	tenantNameToID       map[string]string
	deviceID             string
	scheduledTaskID      string
	evaluationRuleID     string
	updatedSchedule      string
	userID               string
	require              *require.Assertions
	t                    godog.TestingT
}

func NewFeatureContext() *FeatureContext {
	baseURL := "http://localhost:3000"

	if externalURL := os.Getenv("EXTERNAL_API_URL"); externalURL != "" {
		baseURL = externalURL
	}

	return &FeatureContext{
		apiDriver:      driver.NewAPIDriver(baseURL),
		tenantNameToID: make(map[string]string),
	}
}

func IsExternalMode() bool {
	return os.Getenv("EXTERNAL_API_URL") != ""
}

func (fc *FeatureContext) RegisterSteps(ctx *godog.ScenarioContext) {
	// Generic steps
	ctx.Step(`^wait for (.*)$`, fc.waitForDuration)
	ctx.Then(`^the response status code should be (\d+)$`, fc.theResponseStatusCodeShouldBe)
	ctx.Then(`^the response should contain the tenant details$`, fc.theResponseShouldContainTheTenantDetails)
	ctx.Then(`^the response should contain the device details$`, fc.theResponseShouldContainTheDeviceDetails)
	ctx.Then(`^the response should contain the evaluation rule details$`, fc.theResponseShouldContainTheEvaluationRuleDetails)
	ctx.Then(`^the tenant should be soft deleted$`, fc.theTenantShouldBeSoftDeleted)

	// Healthz endpoint steps
	ctx.When(`^I call the healthz endpoint$`, fc.iCallTheHealthzEndpoint)
	ctx.Then(`^the response should contain status information$`, fc.theResponseShouldContainStatusInformation)
	ctx.Then(`^the response should contain version information$`, fc.theResponseShouldContainVersionInformation)
	ctx.Then(`^the response should contain commit hash information$`, fc.theResponseShouldContainCommitHashInformation)

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
	ctx.Then(`^the response should contain the task details$`, fc.theResponseShouldContainTheTaskDetails)
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
	ctx.Then(`^the response should contain the scheduled task details$`, fc.theResponseShouldContainTheScheduledTaskDetails)

	// Interval-based Scheduled Task steps
	ctx.When(`^I create a scheduled task with:$`, fc.iCreateAScheduledTaskWith)
	ctx.Then(`^the response should contain the scheduled task details with interval scheduling$`, fc.theResponseShouldContainTheScheduledTaskDetailsWithIntervalScheduling)
	ctx.Then(`^the response should contain the scheduled task with next execution time$`, fc.theResponseShouldContainTheScheduledTaskWithNextExecutionTime)
	ctx.Then(`^the response should contain the scheduled task details with 3-day interval$`, fc.theResponseShouldContainTheScheduledTaskDetailsWith3DayInterval)
	ctx.Then(`^the response should contain an error about missing initial_day$`, fc.theResponseShouldContainAnErrorAboutMissingInitialDay)
	ctx.When(`^I update the scheduled task with:$`, fc.iUpdateTheScheduledTaskWith)
	ctx.Then(`^the response should contain the updated scheduled task with interval scheduling$`, fc.theResponseShouldContainTheUpdatedScheduledTaskWithIntervalScheduling)

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

	// Tenant Configuration steps
	ctx.Given(`^I have a tenant with id "([^"]*)"$`, fc.iHaveATenantWithIdForConfiguration)
	ctx.When(`^I create a tenant configuration for tenant "([^"]*)" with timezone "([^"]*)"$`, fc.iCreateATenantConfigurationForTenantWithTimezone)
	ctx.When(`^I get the tenant configuration for tenant "([^"]*)"$`, fc.iGetTheTenantConfigurationForTenant)
	ctx.When(`^I update the tenant configuration for tenant "([^"]*)" with timezone "([^"]*)"$`, fc.iUpdateTheTenantConfigurationForTenantWithTimezone)
	ctx.When(`^I update the tenant configuration for tenant "([^"]*)" with timezone "([^"]*)" and version (\d+)$`, fc.iUpdateTheTenantConfigurationForTenantWithTimezoneAndVersion)
	ctx.Given(`^I have a tenant configuration for tenant "([^"]*)" with timezone "([^"]*)"$`, fc.iHaveATenantConfigurationForTenantWithTimezone)
	ctx.Then(`^the response should be "([^"]*)"$`, fc.theResponseShouldBe)
	ctx.Then(`^the error message should be "([^"]*)"$`, fc.theErrorMessageShouldBe)
	ctx.Then(`^the response should contain timezone "([^"]*)"$`, fc.theResponseShouldContainTimezone)
	ctx.Then(`^the tenant configuration should be created successfully$`, fc.theTenantConfigurationShouldBeCreatedSuccessfully)
	ctx.Then(`^the tenant configuration should be retrieved successfully$`, fc.theTenantConfigurationShouldBeRetrievedSuccessfully)
	ctx.Then(`^the tenant configuration should be updated successfully$`, fc.theTenantConfigurationShouldBeUpdatedSuccessfully)

	// User steps
	ctx.When(`^I associate user "([^"]*)" with tenants$`, fc.iAssociateUserWithTenants)
	ctx.Given(`^user "([^"]*)" is associated with tenants$`, fc.userIsAssociatedWithTenants)
	ctx.When(`^I get the user "([^"]*)"$`, fc.iGetTheUser)
	ctx.Then(`^the response should contain the user with id "([^"]*)"$`, fc.theResponseShouldContainTheUserWithId)
	ctx.Then(`^the response should contain exactly (\d+) tenants$`, fc.theResponseShouldContainExactlyTenants)
	ctx.When(`^I update user "([^"]*)" with different tenants$`, fc.iUpdateUserWithDifferentTenants)
	ctx.Given(`^user "([^"]*)" is associated with (\d+) tenants$`, fc.userIsAssociatedWithTenantsCount)
	ctx.When(`^I associate user "([^"]*)" with empty tenant list$`, fc.iAssociateUserWithEmptyTenantList)
	ctx.When(`^I attempt to associate user "([^"]*)" with non-existent tenant$`, fc.iAttemptToAssociateUserWithNonExistentTenant)
	ctx.When(`^I attempt to associate user "([^"]*)" with mixed tenant list$`, fc.iAttemptToAssociateUserWithMixedTenantList)
	ctx.Given(`^I have a user "([^"]*)" associated with tenant "([^"]*)"$`, fc.iHaveAUserAssociatedWithTenant)
	ctx.Given(`^another tenant exists with name "([^"]*)" and email "([^"]*)"$`, fc.anotherTenantExistsWithNameAndEmail)
	ctx.Given(`^a third tenant exists with name "([^"]*)" and email "([^"]*)"$`, fc.aThirdTenantExistsWithNameAndEmail)


	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		fc.t = godog.T(ctx)
		fc.require = require.New(fc.t)

		fc.reset()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		fc.cleanupWebSocket()
		if !IsExternalMode() {
			fc.sendSIGHUPToServer()
		}
		return ctx, err
	})
}

func (fc *FeatureContext) reset() {
	fc.response = nil
	fc.responseData = nil
	fc.responseListData = nil
	fc.tenantID = ""
	fc.tenantIDs = nil
	fc.tenantNameToID = make(map[string]string)
	fc.deviceID = ""
	fc.scheduledTaskID = ""
	fc.evaluationRuleID = ""
	fc.updatedSchedule = ""
	fc.userID = ""
}

func (fc *FeatureContext) sendSIGHUPToServer() {
	serverPID := os.Getenv("SERVER_PID")
	if serverPID == "" {
		return
	}

	var pid int
	if _, err := fmt.Sscanf(serverPID, "%d", &pid); err != nil {
		return
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return
	}

	if err := process.Signal(syscall.SIGHUP); err != nil {
		return
	}

	time.Sleep(10 * time.Millisecond)
}

func (fc *FeatureContext) decodeBody(body io.ReadCloser, target any) error {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bodyBytes, target)
}

func (fc *FeatureContext) decodePaginatedResponse(body *http.Response) ([]map[string]any, error) {
	var paginatedResp PaginatedResponse[map[string]any]
	if err := fc.decodeBody(body.Body, &paginatedResp); err != nil {
		return nil, fmt.Errorf("failed to decode paginated response: %w", err)
	}
	return paginatedResp.Data, nil
}
