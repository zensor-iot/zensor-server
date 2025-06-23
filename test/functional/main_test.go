package functional

import (
	"fmt"
	"os"
	"testing"
	"zensor-server/test/functional/steps"

	"github.com/cucumber/godog"
	"github.com/spf13/pflag"
)

var opts = godog.Options{}

func init() {
	godog.BindCommandLineFlags("godog.", &opts)
}

func TestMain(m *testing.M) {
	pflag.Parse()

	featureContext := steps.NewFeatureContext()

	status := godog.TestSuite{
		Name:                 "godogs",
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
		fmt.Println("Before suite")
	})
}
