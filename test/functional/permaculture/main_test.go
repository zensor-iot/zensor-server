package permaculture

import (
	"log/slog"
	"os"
	"testing"
	"zensor-server/test/functional/permaculture/steps"

	"github.com/cucumber/godog"
	"github.com/spf13/pflag"
)

var opts = godog.Options{}

func init() {
	godog.BindCommandLineFlags("godog.", &opts)
}

func TestMain(m *testing.M) {
	pflag.Parse()

	apiURL := steps.BaseURL()
	if steps.IsExternalMode() {
		slog.Info("Running tests against external API", slog.String("api_url", apiURL))
	} else {
		slog.Info("Running tests against local server", slog.String("api_url", apiURL))
	}

	featureContext := steps.NewFeatureContext()

	status := godog.TestSuite{
		Name:                 "permaculture",
		TestSuiteInitializer: InitializeTestSuite,
		ScenarioInitializer:  featureContext.RegisterSteps,
		Options:              &opts,
	}.Run()

	if st := m.Run(); st > status {
		status = st
	}

	os.Exit(status)
}

func InitializeTestSuite(suite *godog.TestSuiteContext) {
	suite.BeforeSuite(func() {
		if steps.IsExternalMode() {
			slog.Info("Running tests in external mode - skipping local setup")
		} else {
			slog.Info("Running tests in local mode")
		}
	})
}
