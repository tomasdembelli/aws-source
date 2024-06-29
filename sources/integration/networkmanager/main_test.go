package networkmanager

import (
	"fmt"
	"os"
	"testing"

	"github.com/overmindtech/aws-source/sources/integration"
)

func TestMain(m *testing.M) {
	if integration.ShouldRunIntegrationTests() {
		fmt.Println("Running integration tests")
		os.Exit(m.Run())
	} else {
		fmt.Println("Skipping integration tests, set RUN_INTEGRATION_TESTS=true to run them")
		os.Exit(0)
	}
}
