package integration

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

func shouldRunIntegrationTests(t *testing.T, resourceID string) {
	runAll, found := os.LookupEnv("RUN_ALL_INTEGRATION_TESTS")
	if found {
		shouldRunAll, err := strconv.ParseBool(runAll)
		if err != nil {
			t.Skipf("failed to parse RUN_ALL_INTEGRATION_TESTS")
			return
		}

		if shouldRunAll {
			return
		} else {
			t.Skipf("skipping integration tests.. set RUN_ALL_INTEGRATION_TESTS=true or individual RUN_%s_INTEGRATION_TESTS=true to run them", strings.ToUpper(resourceID))
		}
	}

	runResource, found := os.LookupEnv(fmt.Sprintf("RUN_%s_INTEGRATION_TESTS", strings.ToUpper(resourceID)))
	if found {
		shouldRunResource, err := strconv.ParseBool(runResource)
		if err != nil {
			t.Skipf("failed to parse RUN_%s_INTEGRATION_TESTS", strings.ToUpper(resourceID))
			return
		}

		if shouldRunResource {
			return
		} else {
			t.Skipf("skipping integration tests.. set RUN_ALL_INTEGRATION_TESTS=true or individual RUN_%s_INTEGRATION_TESTS=true to run them", strings.ToUpper(resourceID))
		}
	}

	t.Skipf("skipping integration tests.. set RUN_ALL_INTEGRATION_TESTS=true or individual RUN_%s_INTEGRATION_TESTS=true to run them", strings.ToUpper(resourceID))
}
