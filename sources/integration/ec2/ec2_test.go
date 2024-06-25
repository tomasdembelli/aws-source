package ec2

import (
	"context"
	"github.com/overmindtech/aws-source/sources/integration"
	"log/slog"
	"testing"
)

func TestIntegrationEC2(t *testing.T) {
	integration.ShouldRunIntegrationTests(t)

	ctx := context.Background()
	logger := slog.Default()

	ec2Client, err := createEC2Client(ctx)
	if err != nil {
		t.Fatalf("Failed to create EC2 client: %v", err)

	}

	t.Run("Setup", func(t *testing.T) {
		if err := setup(ctx, logger, ec2Client); err != nil {
			t.Fatalf("Failed to setup EC2 integration tests: %v", err)
		}
	})

	t.Run("Test EC2", func(t *testing.T) {
		t.Logf("Running EC2 integration tests")
		TestInstanceSource(t)
	})

	t.Run("Teardown", func(t *testing.T) {
		if err := teardown(ctx, logger, ec2Client); err != nil {
			t.Fatalf("Failed to teardown EC2 integration tests: %v", err)
		}
	})
}
