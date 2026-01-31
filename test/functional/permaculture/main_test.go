package permaculture

import (
	"fmt"
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

	var apiURL string
	if externalURL := os.Getenv("EXTERNAL_API_URL"); externalURL != "" {
		apiURL = externalURL
		fmt.Printf("🌍 Running tests against external API: %s\n", apiURL)
	} else {
		apiURL = "http://127.0.0.1:3000"
		fmt.Printf("🏠 Running tests against local server: %s\n", apiURL)
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
			fmt.Println("Running tests in external mode - skipping local setup")
		} else {
			fmt.Println("Running tests in local mode")
		}
	})
}
