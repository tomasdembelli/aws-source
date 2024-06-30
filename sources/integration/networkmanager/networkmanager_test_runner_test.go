package networkmanager

import (
	"context"
	"log/slog"
	"testing"
)

func TestIntegrationNetworkManager(t *testing.T) {
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

	t.Run("Test Network Manager", func(t *testing.T) {
		t.Logf("Running NetworkManager integration tests")
		TestNetworkManager(t)
	})

	t.Run("Teardown", func(t *testing.T) {
		if err := teardown(ctx, logger, networkmanagerClient); err != nil {
			t.Fatalf("Failed to teardown NetworkManager integration tests: %v", err)
		}
	})
}
