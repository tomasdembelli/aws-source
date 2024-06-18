package networkmanager

import (
	"context"
	"github.com/overmindtech/aws-source/sources/integration"
	"testing"
)

func TestIntegrationNetworkManager(t *testing.T) {
	integration.ShouldRunIntegrationTests(t)

	ctx := context.Background()

	t.Run("Setup", func(t *testing.T) {
		if err := setup(ctx); err != nil {
			t.Fatalf("Failed to setup NetworkManager integration tests: %v", err)
		}
	})

	t.Run("TestSomeSource", func(t *testing.T) {
		t.Logf("Running NetworkManager integration tests")
		TestGlobalNetworkSource(t)
	})

	t.Run("Teardown", func(t *testing.T) {
		if err := integration.Teardown(ctx, integration.TagFilter(integration.NetworkManager)); err != nil {
			t.Fatalf("Failed to teardown NetworkManager integration tests: %v", err)
		}
	})
}
