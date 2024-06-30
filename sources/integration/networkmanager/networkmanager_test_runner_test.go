package networkmanager

import (
	"context"
	"log/slog"
	"testing"
)

func TestIntegrationNetworkManager(t *testing.T) {
	TestSetup(t)

	TestNetworkManager(t)

	TestTeardown(t)
}

func TestSetup(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	networkmanagerClient, err := createNetworkManagerClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create NetworkManager client: %v", err)
	}

	if err := setup(ctx, logger, networkmanagerClient); err != nil {
		t.Fatalf("Failed to setup NetworkManager integration tests: %v", err)
	}
}

func TestTeardown(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()

	networkmanagerClient, err := createNetworkManagerClient(ctx)
	if err != nil {
		t.Fatalf("Failed to create NetworkManager client: %v", err)
	}

	if err := teardown(ctx, logger, networkmanagerClient); err != nil {
		t.Fatalf("Failed to teardown NetworkManager integration tests: %v", err)
	}
}
