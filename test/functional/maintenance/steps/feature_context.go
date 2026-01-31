package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
	"zensor-server/test/functional/maintenance/driver"

	"github.com/cucumber/godog"
	"github.com/stretchr/testify/require"
)

type PaginatedResponse[T any] struct {
	Data       []T `json:"data"`
	Pagination struct {
		Total  int `json:"total"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"pagination"`
}

type FeatureContext struct {
	apiDriver               *driver.APIDriver
	response                *http.Response
	responseData            map[string]any
	responseListData        []map[string]any
	tenantID                string
	tenantIDs               []string
	tenantNameToID          map[string]string
	maintenanceActivityID   string
	maintenanceExecutionID  string
	maintenanceExecutionIDs []string
	require                 *require.Assertions
	t                       godog.TestingT
}

func NewFeatureContext() *FeatureContext {
	baseURL := "http://127.0.0.1:3000"

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
	ctx.Step(`^wait for (.*)$`, fc.waitForDuration)
	ctx.Then(`^the response status code should be (\d+)$`, fc.theResponseStatusCodeShouldBe)

	ctx.Given(`^a tenant exists with name "([^"]*)" and email "([^"]*)"$`, fc.aTenantExistsWithNameAndEmail)

	ctx.When(`^I create a maintenance activity for tenant with type "([^"]*)" and name "([^"]*)"$`, fc.iCreateAMaintenanceActivityForTenantWithTypeAndName)
	ctx.When(`^I create a maintenance activity for tenant with custom type "([^"]*)" and name "([^"]*)"$`, fc.iCreateAMaintenanceActivityForTenantWithCustomTypeAndName)
	ctx.Given(`^a maintenance activity exists for tenant with type "([^"]*)" and name "([^"]*)"$`, fc.aMaintenanceActivityExistsForTenantWithTypeAndName)
	ctx.Given(`^a deactivated maintenance activity exists for tenant with type "([^"]*)" and name "([^"]*)"$`, fc.aDeactivatedMaintenanceActivityExistsForTenantWithTypeAndName)
	ctx.When(`^I list all maintenance activities for the tenant$`, fc.iListAllMaintenanceActivitiesForTheTenant)
	ctx.Then(`^the list should contain the maintenance activity with name "([^"]*)"$`, fc.theListShouldContainTheMaintenanceActivityWithName)
	ctx.When(`^I get the maintenance activity by its ID$`, fc.iGetTheMaintenanceActivityByItsID)
	ctx.Then(`^the response should contain the maintenance activity with name "([^"]*)"$`, fc.theResponseShouldContainTheMaintenanceActivityWithName)
	ctx.Then(`^the response should contain the maintenance activity details$`, fc.theResponseShouldContainTheMaintenanceActivityDetails)
	ctx.When(`^I update the maintenance activity with name "([^"]*)"$`, fc.iUpdateTheMaintenanceActivityWithName)
	ctx.When(`^I activate the maintenance activity$`, fc.iActivateTheMaintenanceActivity)
	ctx.When(`^I deactivate the maintenance activity$`, fc.iDeactivateTheMaintenanceActivity)
	ctx.Then(`^the response should contain an active maintenance activity$`, fc.theResponseShouldContainAnActiveMaintenanceActivity)
	ctx.Then(`^the response should contain an inactive maintenance activity$`, fc.theResponseShouldContainAnInactiveMaintenanceActivity)
	ctx.When(`^I delete the maintenance activity$`, fc.iDeleteTheMaintenanceActivity)

	ctx.Given(`^there are (\d+) maintenance executions for the activity$`, fc.thereAreMaintenanceExecutionsForTheActivity)
	ctx.Given(`^a maintenance execution exists for the activity$`, fc.aMaintenanceExecutionExistsForTheActivity)
	ctx.Given(`^an overdue maintenance execution exists for the activity$`, fc.anOverdueMaintenanceExecutionExistsForTheActivity)
	ctx.When(`^I list all maintenance executions for the activity$`, fc.iListAllMaintenanceExecutionsForTheActivity)
	ctx.Then(`^I should receive (\d+) executions$`, fc.iShouldReceiveExecutions)
	ctx.When(`^I get the maintenance execution by its ID$`, fc.iGetTheMaintenanceExecutionByItsID)
	ctx.Then(`^the response should contain the maintenance execution details$`, fc.theResponseShouldContainTheMaintenanceExecutionDetails)
	ctx.When(`^I mark the maintenance execution as completed by "([^"]*)"$`, fc.iMarkTheMaintenanceExecutionAsCompletedBy)
	ctx.Then(`^the response should contain a completed maintenance execution$`, fc.theResponseShouldContainACompletedMaintenanceExecution)
	ctx.Then(`^the response should contain completed_by "([^"]*)"$`, fc.theResponseShouldContainCompletedBy)
	ctx.Then(`^the response should contain an overdue maintenance execution$`, fc.theResponseShouldContainAnOverdueMaintenanceExecution)

	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		fc.t = godog.T(ctx)
		fc.require = require.New(fc.t)

		fc.reset()
		return ctx, nil
	})

	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if !IsExternalMode() {
			fc.sendSIGHUPToServer()
		}
		return ctx, err
	})
}

func (fc *FeatureContext) waitForDuration(duration string) error {
	duration = strings.TrimSpace(duration)

	var d time.Duration

	if strings.HasSuffix(duration, "ms") {
		msStr := strings.TrimSuffix(duration, "ms")
		ms, err := strconv.Atoi(msStr)
		if err != nil {
			return err
		}
		d = time.Duration(ms) * time.Millisecond
	} else if strings.HasSuffix(duration, "s") {
		sStr := strings.TrimSuffix(duration, "s")
		s, err := strconv.Atoi(sStr)
		if err != nil {
			return err
		}
		d = time.Duration(s) * time.Second
	} else {
		ms, err := strconv.Atoi(duration)
		if err != nil {
			return err
		}
		d = time.Duration(ms) * time.Millisecond
	}

	time.Sleep(d)
	return nil
}

func (fc *FeatureContext) theResponseStatusCodeShouldBe(code int) error {
	fc.require.Equal(code, fc.response.StatusCode, "Unexpected status code")
	return nil
}

func (fc *FeatureContext) aTenantExistsWithNameAndEmail(name, email string) error {
	resp, err := fc.apiDriver.CreateTenant(name, email, "A test tenant")
	fc.require.NoError(err)

	if resp.StatusCode == http.StatusConflict {
		listResp, err := fc.apiDriver.ListTenants()
		fc.require.NoError(err)
		fc.require.Equal(http.StatusOK, listResp.StatusCode)

		var listData struct {
			Data []map[string]any `json:"data"`
		}
		err = fc.decodeBody(listResp.Body, &listData)
		fc.require.NoError(err)

		for _, tenant := range listData.Data {
			if tenant["name"] == name {
				fc.tenantID = tenant["id"].(string)
				return nil
			}
		}
		fc.require.Fail("Tenant with name " + name + " not found in list")
		return nil
	}

	fc.require.Equal(http.StatusCreated, resp.StatusCode)

	var data map[string]any
	err = fc.decodeBody(resp.Body, &data)
	fc.require.NoError(err)

	tenantID := data["id"].(string)
	fc.tenantID = tenantID
	fc.tenantIDs = append(fc.tenantIDs, tenantID)

	time.Sleep(50 * time.Millisecond)
	return nil
}

func (fc *FeatureContext) reset() {
	fc.response = nil
	fc.responseData = nil
	fc.responseListData = nil
	fc.tenantID = ""
	fc.tenantIDs = nil
	fc.tenantNameToID = make(map[string]string)
	fc.maintenanceActivityID = ""
	fc.maintenanceExecutionID = ""
	fc.maintenanceExecutionIDs = nil
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
