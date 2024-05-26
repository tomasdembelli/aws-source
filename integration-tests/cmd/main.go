package main

import (
	"flag"
	"log/slog"
	"os"
	"strconv"
	"testing"

	ec2integration "github.com/overmindtech/aws-source/sources/ec2/integration"
	ecsintegration "github.com/overmindtech/aws-source/sources/ecs/integration"
)

type TestRunner struct {
	setup       func() error
	testsByName map[string]func(*testing.T)
	teardown    func(logger *slog.Logger) error
}

var testRunnersByResourceGroup = map[string]TestRunner{
	"ec2": {
		setup:       ec2integration.Setup,
		testsByName: map[string]func(*testing.T){"TestInstanceSource": ec2integration.TestInstanceSource},
		teardown:    ec2integration.Teardown,
	},
	"ecs": {
		setup:       ecsintegration.Setup,
		testsByName: map[string]func(*testing.T){"TestServiceSource": ecsintegration.TestServiceSource},
		teardown:    ecsintegration.Teardown,
	},
}

func main() {
	action := flag.String("action", "", "action to perform (setup, run, teardown)")
	resourceGroup := flag.String("resource-group", "", "resource group name to run tests for, i.e., 'ec2'")
	flag.Parse()

	logger := slog.Default()

	if action == nil || *action == "" {
		logger.Error("expected 'setup', 'run' or 'teardown' subcommands")
		os.Exit(1)
	}

	if resourceGroup == nil || *resourceGroup == "" {
		// TODO: support `all` as a special value
		// TODO: support multiple resource groups
		logger.Error("expected 'resource-group' flag, i.e., 'ec2'")
		os.Exit(1)
	}

	if !shouldRunIntegrationTests() {
		logger.Warn("skipping integration tests.. set RUN_INTEGRATION_TESTS=true to run them")
		os.Exit(0)
	}

	testRunner, found := testRunnersByResourceGroup[*resourceGroup]
	if !found {
		logger.Error("unknown resource group", slog.String("name", *resourceGroup))
		os.Exit(1)
	}

	switch *action {
	case "setup":
		if err := testRunner.setup(); err != nil {
			logger.Error(
				"failed to setup integration test environment",
				slog.String("name", *resourceGroup),
				slog.String("err", err.Error()),
			)

			os.Exit(1)
		}

		os.Exit(0)
	case "run":
		for testName, testFunc := range testRunner.testsByName {
			code := testing.RunTests(
				func(pattern string, str string) (bool, error) {
					return true, nil
				},
				[]testing.InternalTest{{testName, testFunc}},
			)

			if !code {
				logger.Error("test failed", slog.String("name", testName))
				os.Exit(1)
			}
		}

		os.Exit(0)
	case "teardown":
		if err := testRunner.teardown(logger); err != nil {
			logger.Error(
				"failed to teardown integration test environment",
				slog.String("name", *resourceGroup),
				slog.String("err", err.Error()),
			)

			os.Exit(1)
		}

		os.Exit(0)
	default:
		logger.Error("unknown action", slog.String("action", *action))
		os.Exit(1)
	}
}

func shouldRunIntegrationTests() bool {
	envVar, found := os.LookupEnv("RUN_INTEGRATION_TESTS")
	if !found {
		return false
	}

	// Parse the environment variable into a boolean
	boolVar, err := strconv.ParseBool(envVar)
	if err != nil {
		return false
	}

	return boolVar
}
