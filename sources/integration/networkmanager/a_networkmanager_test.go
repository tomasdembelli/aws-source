package networkmanager

import (
	"context"
	"github.com/overmindtech/aws-source/sources/integration"
	"log/slog"
	"testing"
)

func TestIntegrationNetworkManager(t *testing.T) {
	integration.ShouldRunIntegrationTests(t)

	ctx := context.Background()
	logger := slog.Default()

	networkmanagerClient, err := createNetworkManagerClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create NetworkManager client: %v", err)
	}

	t.Run("Setup", func(t *testing.T) {
		if err := setup(ctx, logger, networkmanagerClient); err != nil {
			t.Fatalf("Failed to setup NetworkManager integration tests: %v", err)
		}
	})

	t.Run("Test Global Network", func(t *testing.T) {
		t.Logf("Running NetworkManager integration tests")
		TestGlobalNetworkSource(t)
	})

	t.Run("Teardown", func(t *testing.T) {
		if err := teardown(ctx, logger, networkmanagerClient); err != nil {
			t.Fatalf("Failed to teardown NetworkManager integration tests: %v", err)
		}
	})
}
